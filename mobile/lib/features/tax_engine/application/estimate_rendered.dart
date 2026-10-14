import 'package:control_contable/api/generated/models/taxpayer_estimate.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Told each time the estimate screen shows a loaded [estimate] (opened,
/// another month chosen, refreshed). It must return at once and never
/// throw: the screen does not wait for it.
typedef EstimateRenderedListener = void Function(TaxpayerEstimate estimate);

/// Who hears that an estimate was shown. The tax engine knows no listener:
/// the app root wires one (History marks the month viewed, M8), so this
/// slice never imports another. Nobody listens by default.
final estimateRenderedListenerProvider = Provider<EstimateRenderedListener>(
  (ref) => _nobodyListens,
  name: 'estimateRenderedListenerProvider',
);

void _nobodyListens(TaxpayerEstimate estimate) {}
