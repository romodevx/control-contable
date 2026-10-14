import 'dart:math';

import 'package:control_contable/features/taxpayers/application/client_id.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

final RegExp _uuidV4 = RegExp(
  r'^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$',
);

void main() {
  test('a version 4, RFC 9562 variant UUID', () {
    for (var seed = 0; seed < 50; seed++) {
      expect(newUuidV4(Random(seed)), matches(_uuidV4));
    }
  });

  test('a secure source by default, a new id each time', () {
    expect(newUuidV4(), isNot(newUuidV4()));
  });

  test('the provider makes ids', () {
    final container = ProviderContainer.test();

    expect(container.read(clientIdProvider)(), matches(_uuidV4));
  });
}
