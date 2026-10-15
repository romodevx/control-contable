import 'package:control_contable/core/records/record_month.dart';

const String _scheme = 'controlcontable';
const String _estimateHost = 'estimate';
const String _periodParameter = 'period';

/// The `YYYY-MM` period a reminder's deep link opens
/// (`controlcontable://estimate?period=YYYY-MM`, MODULE-reminders §2.5),
/// or `null` for anything else. The link comes from the push payload, so
/// it is parsed strictly: an unknown host or a malformed period opens Home.
String? estimatePeriodOfLink(String link) {
  final uri = Uri.tryParse(link);
  if (uri == null || uri.scheme != _scheme || uri.host != _estimateHost) {
    return null;
  }
  final period = uri.queryParameters[_periodParameter];
  return period != null && isMonthPeriod(period) ? period : null;
}
