/// Lowercase letters the synthetic password cycles through.
const int _firstLetter = 0x61; // 'a'
const int _alphabetSize = 26;

/// The length the sign-up form accepts comfortably (contract minimum: 12).
const int syntheticPasswordLength = 16;

/// A password-shaped test value built at run time: `abcdefghijklmnop` for the
/// default [length]. This repository is public, and secret scanners flag
/// credential-like string literals in tests, so no test spells one out.
/// Deterministic: the same [length] always yields the same value.
String syntheticPassword([int length = syntheticPasswordLength]) =>
    String.fromCharCodes([
      for (var i = 0; i < length; i++) _firstLetter + i % _alphabetSize,
    ]);
