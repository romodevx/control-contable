import 'package:control_contable/features/tax_engine/application/estimate.dart';
import 'package:control_contable/features/tax_engine/domain/prior_payment_suggestion.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_riverpod/misc.dart';

/// Whose suggestion: a taxpayer, a regime and a month (`YYYY-MM`).
typedef PriorPaymentQuery = ({String taxpayerId, String regime, String period});

/// The app's estimate of one month's provisional payment, in centavos (see
/// [suggestedPriorPayment]), read through that month's estimate: one
/// request per month, only for empty months on screen. UX only: a failure
/// offers nothing and never blocks the form.
final FutureProviderFamily<int?, PriorPaymentQuery>
priorPaymentSuggestionProvider = FutureProvider.autoDispose
    .family<int?, PriorPaymentQuery>(
      (ref, query) async => suggestedPriorPayment(
        await ref.watch(
          estimateProvider((taxpayerId: query.taxpayerId, period: query.period))
              .future,
        ),
        query.regime,
      ),
      name: 'priorPaymentSuggestionProvider',
      retry: (retryCount, error) => null,
    );
