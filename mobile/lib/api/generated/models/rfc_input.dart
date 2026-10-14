// coverage:ignore-file
// GENERATED CODE - DO NOT MODIFY BY HAND
// ignore_for_file: type=lint, unused_import, invalid_annotation_target, unnecessary_import

/// RFC as typed: spaces, dashes and lower case are accepted and normalized server-side (trim, upper-case, NFC, inner spaces and `-` removed). The edge bounds only the length, so every RFC rule is reported on `rfc` with a domain code: `rfc_length_mismatch` (13 for física, 12 for moral), `rfc_format`, `rfc_date`, `rfc_generic`. A SAT módulo 11 check-digit mismatch is not an error (D-RFC): it is accepted with the warning `rfc_check_digit_mismatch`. Never echoed back.
typedef RfcInput = String;
