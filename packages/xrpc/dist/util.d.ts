import { LexXrpcProcedure, LexXrpcQuery } from '@atproto/lexicon';
import { CallOptions, Headers, QueryParams, ResponseType } from './types';
export declare function getMethodSchemaHTTPMethod(schema: LexXrpcProcedure | LexXrpcQuery): "post" | "get";
export declare function constructMethodCallUri(nsid: string, schema: LexXrpcProcedure | LexXrpcQuery, serviceUri: URL, params?: QueryParams): string;
export declare function encodeQueryParam(type: 'string' | 'float' | 'integer' | 'boolean' | 'datetime' | 'array' | 'unknown', value: any): string;
export declare function normalizeHeaders(headers: Headers): Headers;
export declare function constructMethodCallHeaders(schema: LexXrpcProcedure | LexXrpcQuery, data?: any, opts?: CallOptions): Headers;
export declare function encodeMethodCallBody(headers: Headers, data?: any): ArrayBuffer | undefined;
export declare function httpResponseCodeToEnum(status: number): ResponseType;
export declare function httpResponseBodyParse(mimeType: string | null, data: ArrayBuffer | undefined): any;
//# sourceMappingURL=util.d.ts.map