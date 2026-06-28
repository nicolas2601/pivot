<script lang="ts">
  interface Props {
    label: string;
    name: string;
    type?: 'text' | 'email' | 'password' | 'number' | 'date';
    value: string | number;
    placeholder?: string;
    required?: boolean;
    error?: string;
    hint?: string;
    autocomplete?: 'email' | 'current-password' | 'new-password' | 'name' | 'off' | 'on';
    min?: number;
    step?: number;
    maxLength?: number;
    onChange?: (value: string | number) => void;
  }

  let {
    label,
    name,
    type = 'text',
    value = $bindable(''),
    placeholder = '',
    required = false,
    error,
    hint,
    autocomplete,
    min,
    step,
    maxLength,
    onChange
  }: Props = $props();

  // Use $derived for IDs (re-evaluates when name changes)
  const inputId = $derived(`input-${name}`);
  const errorId = $derived(`error-${name}`);
  const hintId = $derived(`hint-${name}`);

  function handleInput(e: Event) {
    const target = e.currentTarget as HTMLInputElement;
    const newValue = type === 'number' ? Number(target.value) : target.value;
    value = newValue as never;
    onChange?.(newValue);
  }
</script>

<div class="space-y-1.5">
  <label for={inputId} class="block text-sm font-medium text-ink">
    {label}
    {#if required}<span class="text-semantic-error">*</span>{/if}
  </label>
  <input
    id={inputId}
    {name}
    {type}
    {value}
    {placeholder}
    {required}
    {autocomplete}
    {min}
    {step}
    maxLength={maxLength}
    aria-invalid={error ? 'true' : undefined}
    aria-describedby={[error && errorId, hint && hintId].filter(Boolean).join(' ') || undefined}
    oninput={handleInput}
    class="w-full px-4 py-3 h-11 bg-surface-card text-ink rounded-md border {error ? 'border-semantic-error' : 'border-hairline-strong'} focus:outline-none focus:border-ink focus:ring-1 focus:ring-ink transition-colors"
  />
  {#if hint && !error}
    <p id={hintId} class="text-xs text-muted">{hint}</p>
  {/if}
  {#if error}
    <p id={errorId} class="text-xs text-semantic-error">{error}</p>
  {/if}
</div>