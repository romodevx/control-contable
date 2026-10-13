import 'package:control_contable/core/config/app_config.dart';
import 'package:control_contable/core/network/api_client_provider.dart';
import 'package:control_contable/core/network/dio_provider.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_test/flutter_test.dart';

import '../../helpers/recording_adapter.dart';

void main() {
  test(
    'the generated client rides on the one dio, per-flavor base URL',
    () async {
      final container = ProviderContainer.test(
        overrides: [
          appConfigProvider.overrideWithValue(
            AppConfig.resolve(flavor: AppFlavor.dev, useMock: true),
          ),
        ],
      );
      final adapter = RecordingAdapter(
        replies: {
          '/v1/flags': (
            status: 200,
            body: {
              'enabled': ['module_auth', 'module_incomes'],
            },
          ),
        },
      );
      container.read(dioProvider).httpClientAdapter = adapter;

      final flags = await container
          .read(apiClientProvider)
          .platform
          .listEnabledFlags();

      expect(flags.enabled, ['module_auth', 'module_incomes']);
      expect(
        adapter.requests.single.uri.toString(),
        'http://localhost:3001/v1/flags',
      );
    },
  );
}
