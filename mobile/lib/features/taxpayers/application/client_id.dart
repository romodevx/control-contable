import 'dart:math';

import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Bytes in a UUID.
const int _uuidBytes = 16;
const int _byteValues = 256;

/// Byte 6 carries the version nibble; byte 8 the RFC 9562 variant bits.
const int _versionByte = 6;
const int _variantByte = 8;
const int _version4 = 0x40;
const int _variantRfc = 0x80;
const int _lowNibble = 0x0f;
const int _lowSixBits = 0x3f;

/// Where the hyphens go, as offsets into the 32 hex digits.
const List<int> _hyphenAfter = [8, 12, 16, 20];

/// A random (version 4) UUID from [random], a secure source by default.
String newUuidV4([Random? random]) {
  final source = random ?? Random.secure();
  final bytes = [
    for (var i = 0; i < _uuidBytes; i++) source.nextInt(_byteValues),
  ];
  bytes[_versionByte] = (bytes[_versionByte] & _lowNibble) | _version4;
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

/// Makes the client id of a create (`TaxpayerCreate.id`), so a retry after
/// a lost answer is idempotent. Tests override it with a fixed id.
final clientIdProvider = Provider<String Function()>(
  (ref) => newUuidV4,
  name: 'clientIdProvider',
);
