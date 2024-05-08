import { LexiconDoc, Lexicons } from '@atproto/lexicon';
import { FetchHandler, FetchFunction, FetchHandlerResponse, Headers, CallOptions, QueryParams, XRPCResponse } from './types';
export declare class Client {
    fetch: FetchHandler;
    lex: Lexicons;
    call(serviceUri: string | URL, methodNsid: string, params?: QueryParams, data?: unknown, opts?: CallOptions): Promise<XRPCResponse>;
    service(serviceUri: string | URL, fetch: FetchFunction): ServiceClient;
    addLexicon(doc: LexiconDoc): void;
    addLexicons(docs: LexiconDoc[]): void;
    removeLexicon(uri: string): void;
}
export declare class ServiceClient {
    baseClient: Client;
    uri: URL;
    headers: Record<string, string>;
    fetch: FetchFunction;
    constructor(baseClient: Client, serviceUri: string | URL, fetch: FetchFunction);
    setHeader(key: string, value: string): void;
    unsetHeader(key: string): void;
    call(methodNsid: string, params?: QueryParams, data?: unknown, opts?: CallOptions): Promise<XRPCResponse>;
}
export declare function defaultFetchHandler(httpUri: string, httpMethod: string, httpHeaders: Headers, httpReqBody: unknown, fetch: FetchFunction): Promise<FetchHandlerResponse>;
//# sourceMappingURL=client.d.ts.map