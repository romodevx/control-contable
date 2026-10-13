import 'package:control_contable/core/links/url_opener.dart';
import 'package:flutter_test/flutter_test.dart';

void main() {
  test(
    'refuses anything that is not a web page, without launching it',
    () async {
      for (final url in [
        'javascript:alert(1)',
        'file:///etc/hosts',
        'intent://scan/#Intent;scheme=zxing;end',
      ]) {
        expect(await openWebPage(Uri.parse(url)), isFalse, reason: url);
      }
    },
  );
}
