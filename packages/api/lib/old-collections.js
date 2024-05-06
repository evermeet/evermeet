export function makeCollections(api) {
  return {
    events: {
      schema: {
        version: 0,
        primaryKey: "id",
        type: "object",
        properties: {
          id: {
            type: "string",
            maxLength: 32,
          },
          slug: {
            type: "string",
          },
          name: {
            type: "string",
          },
          calendarId: {
            type: "string",
          },
          guestsNative: {
            type: "array",
            items: {
              type: "object",
              properties: {
                ref: {
                  type: "string",
                },
                rel: {
                  type: "string",
                },
                time: {
                  type: "string",
                },
              },
            },
          },
        },
        encrypted: ["name"],
      },
      methods: {
        async view(opts = {}) {
          const json = JSON.parse(JSON.stringify(this.toJSON()));
          if (json.calendarId) {
            const calendar = await api.cols.calendars
              .findOne({ selector: { id: json.calendarId } })
              .exec();
            if (calendar && opts.calendar !== false) {
              json.calendar = await calendar.view(opts);
            }
          }
          json.guestCountNative = (json.guestsNative || []).length;
          json.guestCountTotal = json.guestCountNative + (json.guestCount || 0);
          return json;
        },
        async view_audit() {
          return [];
        },
      },
    },
    remoteobjects: {
      schema: {
        version: 0,
        primaryKey: "id",
        type: "object",
        properties: {
          id: {
            type: "string",
            maxLength: 32,
          },
        },
      },
    },
    calendars: {
      schema: {
        version: 0,
        primaryKey: "id",
        type: "object",
        properties: {
          id: {
            type: "string",
            maxLength: 32,
          },
          name: {
            type: "string",
          },
          slug: {
            type: "string",
          },
          personal: {
            type: "boolean",
          },
          managers: {
            type: "array",
            items: {
              type: "object",
              properties: {
                ref: {
                  type: "string",
                },
                time: {
                  type: "string",
                },
              },
            },
          },
        },
      },
      methods: {
        async view(opts = {}) {
          const json = JSON.parse(JSON.stringify(this.toJSON()));
          if (opts.events !== false) {
            json.events = [];
            const events = await api.cols.events
              .find({ selector: { calendarId: json.id } })
              .exec();
            for (const e of events) {
              json.events.push(
                await e.view(Object.assign(opts, { calendar: false })),
              );
            }
          }
          return json;
        },
      },
    },
    users: {
      schema: {
        version: 0,
        primaryKey: "id",
        type: "object",
        properties: {
          id: {
            type: "string",
            maxLength: 32,
          },
          did: {
            type: "string",
          },
          name: {
            type: "string",
          },
          emails: {
            type: "array",
            items: {
              type: "string",
              format: "email",
            },
          },
          emailShas: {
            type: "array",
            items: {
              type: "string",
            },
          },
          calendarsManage: {
            type: "array",
            items: {
              type: "object",
              properties: {
                ref: {
                  type: "string",
                },
                time: {
                  type: "string",
                },
              },
            },
          },
          subscribedCalendars: {
            type: "array",
            maxItems: 1000,
            items: {
              type: "object",
              properties: {
                ref: {
                  type: "string",
                },
                time: {
                  type: "string",
                },
              },
            },
          },
          personalCalendar: {
            type: "string",
          },
          events: {
            type: "array",
            items: {
              type: "object",
              properties: {
                ref: {
                  type: "string",
                },
                rel: {
                  type: "string",
                },
                time: {
                  type: "string",
                },
              },
            },
          },
        },
        encrypted: ["emails"],
      },
    },
    sessions: {
      schema: {
        version: 0,
        primaryKey: "id",
        type: "object",
        properties: {
          id: {
            type: "string",
            maxLength: 42,
          },
          user: {
            type: "string",
          },
          expiry: {
            type: "string",
          },
          time: {
            type: "string",
          },
        },
      },
    },
  };
}
