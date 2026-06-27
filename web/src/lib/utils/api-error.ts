export interface ApiError {
  code: string;
  message: string;
}

export class ApiException extends Error {
  constructor(
    public status: number,
    public apiError: ApiError
  ) {
    super(apiError.message);
    this.name = 'ApiException';
  }

  get isUnauthorized(): boolean {
    return this.status === 401;
  }

  get isConflict(): boolean {
    return this.status === 409;
  }

  get isValidation(): boolean {
    return this.status === 422;
  }
}