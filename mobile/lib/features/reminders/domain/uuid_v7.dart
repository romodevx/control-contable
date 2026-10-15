import 'dart:math';

const int _uuidBytes = 16;
const int _timestampBytes = 6;
const int _byteValues = 256;
const int _bitsPerByte = 8;

/// Byte 6 carries the version nibble; byte 8 the RFC 9562 variant bits.
const int _versionByte = 6;
const int _variantByte = 8;
const int _version7 = 0x70;
const int _variantRfc = 0x80;
const int _lowNibble = 0x0f;
const int _lowSixBits = 0x3f;
const int _byteMask = 0xff;

/// Where the hyphens go, as offsets into the 32 hex digits.
const List<int> _hyphenAfter = [8, 12, 16, 20];

/// A time-ordered (version 7) UUID for [now]: 48 bits of Unix time in
/// milliseconds, then random bits from [random] (a secure source by
/// default). The installation id of `registerDevice` (MODULE-reminders
/// §2.3).
String newUuidV7(DateTime now, [Random? random]) {
  final source = random ?? Random.secure();
  final millis = now.millisecondsSinceEpoch;
  final bytes = [
    for (var i = 0; i < _timestampBytes; i++)
      (millis >> (_bitsPerByte * (_timestampBytes - 1 - i))) & _byteMask,
    for (var i = _timestampBytes; i < _uuidBytes; i++)
      source.nextInt(_byteValues),
  ];
  bytes[_versionByte] = (bytes[_versionByte] & _lowNibble) | _version7;
  bytes[_variantByte] = (bytes[_variantByte] & _lowSixBits) | _variantRfc;
  final hex = bytes.map((b) => b.toRadixString(16).padLeft(2, '0')).join();
  final buffer = StringBuffer();
  var start = 0;
  for (final end in [..._hyphenAfter, hex.length]) {
    if (start > 0) buffer.write('-');
    buffer.write(hex.substring(start, end));
    start = end;
  }
  return buffer.toString();
}
