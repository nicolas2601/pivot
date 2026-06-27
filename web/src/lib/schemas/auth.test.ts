import { describe, expect, it } from 'vitest';
import { LoginInputSchema, RegisterInputSchema, UserSchema } from './auth';

describe('auth schemas', () => {
  describe('LoginInputSchema', () => {
    it('accepts valid email and password', () => {
      const result = LoginInputSchema.safeParse({ email: 'user@example.com', password: 'any' });
      expect(result.success).toBe(true);
    });

    it('rejects invalid email', () => {
      const result = LoginInputSchema.safeParse({ email: 'not-an-email', password: 'any' });
      expect(result.success).toBe(false);
    });

    it('rejects empty password', () => {
      const result = LoginInputSchema.safeParse({ email: 'user@example.com', password: '' });
      expect(result.success).toBe(false);
    });
  });

  describe('RegisterInputSchema', () => {
    it('accepts password >= 8 chars', () => {
      const result = RegisterInputSchema.safeParse({ email: 'a@b.com', password: '12345678' });
      expect(result.success).toBe(true);
    });

    it('rejects password < 8 chars', () => {
      const result = RegisterInputSchema.safeParse({ email: 'a@b.com', password: 'short' });
      expect(result.success).toBe(false);
      if (!result.success) {
        expect(result.error.issues[0].message).toContain('8 caracteres');
      }
    });

    it('accepts display_name as optional', () => {
      const result = RegisterInputSchema.safeParse({ email: 'a@b.com', password: 'longpassword' });
      expect(result.success).toBe(true);
    });
  });

  describe('UserSchema', () => {
    it('parses valid user', () => {
      const result = UserSchema.safeParse({
        id: '550e8400-e29b-41d4-a716-446655440000',
        email: 'user@example.com',
        display_name: 'Test',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z'
      });
      expect(result.success).toBe(true);
    });

    it('allows null display_name', () => {
      const result = UserSchema.safeParse({
        id: '550e8400-e29b-41d4-a716-446655440000',
        email: 'user@example.com',
        display_name: null,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z'
      });
      expect(result.success).toBe(true);
    });
  });
});