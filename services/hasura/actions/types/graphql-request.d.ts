// Type declarations for graphql-request to satisfy strict TypeScript requirements
declare module 'graphql-request' {
  export type Variables = Record<string, any>;

  export class GraphQLClient {
    constructor(url: string, options?: {
      headers?: Record<string, string>;
      timeout?: number;
      fetch?: typeof fetch;
    });

    request<T = any>(
      document: string | any,
      variables?: Variables
    ): Promise<T>;
  }
  
  export function gql(
    literals: TemplateStringsArray,
    ...placeholders: any[]
  ): string;
  
  export class ClientError extends Error {
    response: {
      status: number;
      headers: Headers;
    };
    request: {
      query: string;
      variables?: Variables;
    };
  }
}
