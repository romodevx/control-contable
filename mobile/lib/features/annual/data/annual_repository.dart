import 'package:control_contable/api/generated/clients/annual_client.dart';
import 'package:control_contable/api/generated/models/annual_breakdown.dart';
import 'package:control_contable/api/generated/models/annual_computed_section_kind.dart';
import 'package:control_contable/api/generated/models/annual_estimate.dart';
import 'package:control_contable/core/network/api_client_provider.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The annual reads of `module_annual` (`/v1/taxpayers/{id}/annual-*`).
/// Every method throws an `ApiFailure` when the server refuses or cannot be
/// reached.
abstract interface class AnnualRepository {
  /// The annual estimate of [taxpayerId] for [fiscalYear].
  Future<AnnualEstimate> getEstimate({
    required String taxpayerId,
    required int fiscalYear,
  });

  /// The computation trace of [section] in [fiscalYear].
  Future<AnnualBreakdown> getBreakdown({
    required String taxpayerId,
    required int fiscalYear,
    required AnnualComputedSectionKind section,
  });
}

/// [AnnualRepository] over the generated client.
class ApiAnnualRepository implements AnnualRepository {
  /// The repository on [_client].
  new(this._client);

  final AnnualClient _client;

  @override
  Future<AnnualEstimate> getEstimate({
    required String taxpayerId,
    required int fiscalYear,
  }) => guardApi(
    () => _client.getAnnualEstimate(
      taxpayerId: taxpayerId,
      fiscalYear: fiscalYear,
    ),
  );

  @override
  Future<AnnualBreakdown> getBreakdown({
    required String taxpayerId,
    required int fiscalYear,
    required AnnualComputedSectionKind section,
  }) => guardApi(
    () => _client.getAnnualBreakdown(
      taxpayerId: taxpayerId,
      fiscalYear: fiscalYear,
      section: section,
    ),
  );
}

/// The annual operations. Tests override them with a fake.
final annualRepositoryProvider = Provider<AnnualRepository>(
  (ref) => ApiAnnualRepository(ref.watch(apiClientProvider).annual),
  name: 'annualRepositoryProvider',
);
