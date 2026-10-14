import 'package:control_contable/api/generated/models/person_type.dart';
import 'package:control_contable/api/generated/models/regime.dart';
import 'package:control_contable/features/taxpayers/data/taxpayers_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';

/// The regime catalog of one person type, read while the regimes step is
/// shown. A failed load shows a retry button: no automatic retry.
final FutureProviderFamily<List<Regime>, PersonType> regimeCatalogProvider =
    FutureProvider.autoDispose.family<List<Regime>, PersonType>(
      (ref, personType) =>
          ref.watch(taxpayersRepositoryProvider).listRegimes(personType),
      name: 'regimeCatalogProvider',
      retry: _noAutomaticRetry,
    );

Duration? _noAutomaticRetry(int retryCount, Object error) => null;
