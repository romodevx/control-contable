import 'package:control_contable/api/generated/models/me.dart';
import 'package:control_contable/api/generated/models/user.dart';
import 'package:control_contable/features/auth/application/profile.dart';
import 'package:flutter_test/flutter_test.dart';

import 'auth_fakes.dart';

void main() {
  test('the profile of Me is its User fields, access left out', () {
    final me = Me(
      id: sampleUser.id,
      email: sampleUser.email,
      name: sampleUser.name,
      createdAt: sampleUser.createdAt,
      access: sampleAccess,
    );

    expect(me.profile, sampleUser);
  });

  test('a nameless Me has a nameless profile', () {
    final me = Me(
      id: sampleUser.id,
      email: sampleUser.email,
      createdAt: sampleUser.createdAt,
      access: sampleAccess,
    );

    expect(me.profile, sampleUser.copyWith(name: null));
    expect(me.profile, isA<User>());
  });
}
