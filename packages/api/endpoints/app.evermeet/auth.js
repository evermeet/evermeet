import argon2 from "argon2";
import jwt from "jsonwebtoken";
import { DidJwk } from "@web5/dids";

import { Secp256k1Keypair } from "@atproto/crypto";
import { Client as PlcClient } from "../../../did-plc-lib";

function base64ToBytes(base64) {
  const binString = atob(base64);
  return Uint8Array.from(binString, (m) => m.codePointAt(0));
}

function bytesToBase64(bytes) {
  const binString = Array.from(bytes, (byte) =>
    String.fromCodePoint(byte),
  ).join("");
  return btoa(binString);
}

function createSessionToken(user, ctx) {
  return jwt.sign({ user: user.handle }, ctx.api.config.api.jwtSecret);
}

// ----------------------------------------------------------------------

export async function generateDid(server) {
  server.endpoint(async (ctx) => {
    const { db, input } = ctx;
    const { handle } = ctx.input;
    const plcClient = new PlcClient(ctx.api.config.plcServer);

    const obj = await ctx.api.objectGet(ctx, handle);
    if (obj) {
      return { error: "HandleNotAvailable" };
    }

    let pdid, done;
    while (!done) {
      const signingKey = await Secp256k1Keypair.create({ exportable: true });
      const rotationKey = await Secp256k1Keypair.create({ exportable: true });

      pdid = await plcClient.createDidLocal({
        signingKey: signingKey.did(),
        handle,
        pds: `https://${ctx.api.config.domain}`,
        rotationKeys: [rotationKey.did()],
        signer: rotationKey,
      });
      try {
        await plcClient.getDocument(pdid.did);
      } catch (e) {
        done = true;
      }
    }

    return {
      body: {
        did: pdid.did,
      },
    };
  });
}

export async function createAccount(server, ctx) {
  server.endpoint(async ({ input: { handle, password }, db }) => {
    const domain = handle.split(".").slice(1).join(".").toLowerCase();

    // check if its available domain on this instance
    if (!ctx.api.config.availableUserDomains.includes("." + domain)) {
      return { error: "UnsupportedDomain" };
    }

    // check if handle is avaiblable
    const exists = await db.users.findOne({ handle });
    if (exists) {
      return { error: "HandleNotAvailable" };
    }

    // TODO additional password check
    const passwordHash = await argon2.hash(password);

    // create DID
    // const didJwk = await DidJwk.create()

    const signingKey = await Secp256k1Keypair.create({ exportable: true });
    const rotationKey = await Secp256k1Keypair.create({ exportable: true });

    const plcClient = new PlcClient(ctx.api.config.plcServer);
    const did = await plcClient.createDid({
      signingKey: signingKey.did(),
      handle,
      pds: `https://${ctx.api.config.domain}`,
      rotationKeys: [rotationKey.did()],
      signer: rotationKey,
    });

    // create user object
    const user = await db.users.create({
      handle,
      did,
      // portableDid: await did.export(),
      password: passwordHash,
      signingKey: bytesToBase64(await signingKey.export()),
      rotationKey: bytesToBase64(await rotationKey.export()),
    });

    // save user to database
    await db.em.persist(user).flush();

    // create session
    const token = createSessionToken(user, ctx);
    const { body: session } = await ctx.api.request(
      "app.evermeet.auth.createSession",
      {
        input: { identifier: handle, password },
      },
    );

    return {
      body: {
        accessToken: session.accessJwt,
        handle,
        did: user.did,
      },
    };
  });
}

export function createSession(server, ctx) {
  server.endpoint(async ({ input, db }) => {
    if (!input) {
      return { error: "InputNotSpecified" };
    }
    let ident = input.identifier;
    const password = input.password;

    if (!ident || !password) {
      return { error: "InputNotSpecified" };
    }
    // normalize input
    ident = ident.toLowerCase();

    // find user
    let user;
    if (ident.match(/@/)) {
      user = await db.users.findOne({ email: ident });
    } else {
      user = await db.users.findOne({ handle: ident });
    }
    if (!user) {
      return { error: "BadCredentials" };
    }

    // verify password
    const verified = await argon2.verify(user.password, password);
    if (!verified) {
      return { error: "BadCredentials" };
    }

    // create JWT token
    const token = createSessionToken(user, ctx);

    // create session
    const session = await db.sessions.create({ userId: user.id, token });
    db.em.persist(session).flush();

    return {
      encoding: "application/json",
      body: {
        accessJwt: token,
        // refreshJwt: 'xxx',
        handle: user.handle,
        did: user.did,
        user: await user.view(),
      },
    };
  });
}

export async function getSession(server, ctx) {
  server.endpoint({
    auth: ctx.api.authVerifier.accessUser,
    handler: async ({ db, user }) => {
      return {
        body: {
          // accessJwt: session.token,
          // refreshJwt: 'xxx',
          handle: user.handle,
          did: user.did,
          user: await user.view(),
        },
      };
    },
  });
}

export async function updateAccount(server, { api: { authVerifier } }) {
  server.endpoint({
    auth: authVerifier.accessUser,
    handler: async (ctx) => {
      const { user, input, db } = ctx;

      if (input.profile) {
        // update blobs
        let avatar;
        if (input.profile.avatar) {
          const blob = await db.blobs.findOne({
            cid: input.profile.avatar.$cid,
          });
          if (!blob) {
            return { error: "InvalidAvatarBlob" };
          }
          avatar = blob.cid;
        }
        console.log(avatar);
        user.name = input.profile.name;
        user.description = input.profile.description;
        user.avatarBlob = avatar;
      }

      if (input.preferences) {
        for (const p of Object.keys(input.preferences)) {
          user.preferences[p] = input.preferences[p];
        }
      }

      await db.em.flush();
      return {
        body: {
          user: await user.view(ctx),
        },
      };
    },
  });
}
