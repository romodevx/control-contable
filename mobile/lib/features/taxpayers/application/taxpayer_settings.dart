import 'package:control_contable/core/access/read_only_access.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/features/taxpayers/application/taxpayer_list.dart';
import 'package:control_contable/features/taxpayers/data/taxpayers_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Whether a write needs a plan (OQ-8, MODULE-taxpayers §2.10).
enum WriteAccess {
  /// Rename, RFC correction, unarchive, add regime, options: `402` for a
  /// read-only user.
  needsPlan,

  /// Archive, delete, end a regime: allowed while read-only.
  always,
}

/// Runs the writes of the taxpayer settings: each one through the
/// repository, then the taxpayer list reloads (the switcher, the settings
/// and the active taxpayer follow it).
///
/// A `402 entitlement_required` is remembered for the session
/// (`readOnlyAccessProvider`); once known, a [WriteAccess.needsPlan] write
/// fails with `ApiEntitlementRequired` without asking the server. UX only:
/// the server's `402` stays the boundary.
class TaxpayerSettings {
  /// The writes over [_ref].
  new(this._ref);

  final Ref _ref;

  /// Whether a [WriteAccess.needsPlan] write is already known to fail.
  bool get knownReadOnly => _ref.read(readOnlyAccessProvider);

  /// Runs [write]; throws its `ApiFailure`.
  Future<T> run<T>(
    WriteAccess access,
    Future<T> Function(TaxpayersRepository repository) write,
  ) async {
    if (access == WriteAccess.needsPlan && knownReadOnly) {
      throw const ApiEntitlementRequired();
    }
    try {
      final result = await write(_ref.read(taxpayersRepositoryProvider));
      await _ref.read(taxpayersProvider.notifier).reload();
      return result;
    } on ApiEntitlementRequired {
      _ref.read(readOnlyAccessProvider.notifier).markReadOnly();
      rethrow;
    }
  }
}

/// See [TaxpayerSettings].
final taxpayerSettingsProvider = Provider<TaxpayerSettings>(
  TaxpayerSettings.new,
  name: 'taxpayerSettingsProvider',
);
