import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// Where the active taxpayer's id is kept on the device. Only the id: it is
/// a non-secret preference (MODULE-taxpayers §2.9). Never the RFC, a name
/// or any credential (those live in the secure store only).
abstract interface class ActiveTaxpayerStore {
  /// The kept id, or `null`.
  Future<String?> read();

  /// Keeps [taxpayerId].
  Future<void> write(String taxpayerId);

  /// Forgets the kept id (sign-out).
  Future<void> clear();
}

/// The preference key of the active taxpayer id.
const String activeTaxpayerKey = 'active_taxpayer_id';

/// [ActiveTaxpayerStore] over the platform preferences.
class PreferencesActiveTaxpayerStore implements ActiveTaxpayerStore {
  /// The store over [_preferences].
  new(this._preferences);

  final SharedPreferencesAsync _preferences;

  @override
  Future<String?> read() => _preferences.getString(activeTaxpayerKey);

  @override
  Future<void> write(String taxpayerId) =>
      _preferences.setString(activeTaxpayerKey, taxpayerId);

  @override
  Future<void> clear() => _preferences.remove(activeTaxpayerKey);
}

/// An [ActiveTaxpayerStore] held in memory (tests).
class InMemoryActiveTaxpayerStore implements ActiveTaxpayerStore {
  /// A store that starts with [id] (none by default).
  new([this.id]);

  /// The kept id.
  String? id;

  @override
  Future<String?> read() async => id;

  @override
  Future<void> write(String taxpayerId) async => id = taxpayerId;

  @override
  Future<void> clear() async => id = null;
}

/// The device's active-taxpayer store. Tests override it.
final activeTaxpayerStoreProvider = Provider<ActiveTaxpayerStore>(
  (ref) => PreferencesActiveTaxpayerStore(SharedPreferencesAsync()),
  name: 'activeTaxpayerStoreProvider',
);
