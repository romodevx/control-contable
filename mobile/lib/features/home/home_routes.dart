import 'package:control_contable/core/router/app_router.dart';
import 'package:control_contable/features/home/presentation/home_screen.dart';
import 'package:go_router/go_router.dart';

/// Routes owned by the home feature.
final List<RouteBase> homeRoutes = [
  GoRoute(
    path: AppRoutes.home,
    builder: (context, state) => const HomeScreen(),
  ),
];
