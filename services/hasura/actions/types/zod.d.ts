// Type declarations for Zod to satisfy strict TypeScript requirements
declare module 'zod' {
  export class ZodType<T> {
    optional(): ZodType<T | undefined>;
    min(min: number, message?: string): this;
    max(max: number, message?: string): this;
    int(message?: string): this;
    positive(message?: string): this;
    regex(regex: RegExp, message?: string): this;
    uuid(message?: string): this;
    refine<U extends T>(
      check: (arg: T) => arg is U,
      message?: string | { message: string }
    ): ZodType<U>;
    refine(
      check: (arg: T) => boolean | Promise<boolean>,
      message?: string | { message: string }
    ): this;
    parse(data: unknown): T;
    safeParse(data: unknown): { success: boolean; data?: T; error?: ZodError };
  }
  
  export class ZodError extends Error {
    issues: Array<{ message: string; path: string[] }>;
  }

  export function string(): ZodType<string>;
  export function number(): ZodType<number>;
  export function boolean(): ZodType<boolean>;
  export function array<T>(schema: ZodType<T>): ZodType<T[]>;
  
  export interface ZodObject<T> extends ZodType<T> {
    shape: Record<string, ZodType<any>>;
  }
  
  export function object<T extends Record<string, ZodType<any>>>(shape: T): ZodObject<{
    [k in keyof T]: T[k] extends ZodType<infer U> ? U : never;
  }>;
  
  export function z_enum<T extends readonly [string, ...string[]]>(values: T): ZodType<T[number]>;
  export function parse<T>(schema: ZodType<T>, data: unknown): T;
}
