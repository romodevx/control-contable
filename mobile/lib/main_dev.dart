import 'package:control_contable/app/bootstrap.dart';
import 'package:control_contable/core/config/app_config.dart';

/// Dev flavor: local API (port 3000), or the Prism mock with USE_MOCK=true.
void main() => bootstrap(AppFlavor.dev);
