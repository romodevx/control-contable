import 'package:control_contable/features/taxpayers/data/active_taxpayer_store.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:shared_preferences_platform_interface/in_memory_shared_preferences_async.dart';
import 'package:shared_preferences_platform_interface/shared_preferences_async_platform_interface.dart';

void main() {
  late SharedPreferencesAsync preferences;
  late PreferencesActiveTaxpayerStore store;

  setUp(() {
    SharedPreferencesAsyncPlatform.instance =
        InMemorySharedPreferencesAsync.empty();
    preferences = SharedPreferencesAsync();
    store = PreferencesActiveTaxpayerStore(preferences);
  });

  test('keeps only the id under one key, and forgets it', () async {
    expect(await store.read(), isNull);

    await store.write('0192f0c4-7b1e-7c3a-9d2e-6f1a2b3c4d5e');

    expect(await preferences.getKeys(), {activeTaxpayerKey});
    expect(await store.read(), '0192f0c4-7b1e-7c3a-9d2e-6f1a2b3c4d5e');

    await store.clear();

    expect(await store.read(), isNull);
  });
}
