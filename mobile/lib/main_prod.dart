import 'package:control_contable/app/bootstrap.dart';
import 'package:control_contable/core/config/app_config.dart';

/// Prod (store) flavor: needs --dart-define=API_BASE_URL=https://...
void main() => bootstrap(AppFlavor.prod);
