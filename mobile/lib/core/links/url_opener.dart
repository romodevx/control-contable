import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:url_launcher/url_launcher.dart';

/// Opens a web page; `false` when it could not be opened.
typedef UrlOpener = Future<bool> Function(Uri url);

/// Schemes the app opens. Anything else (`javascript:`, `file:`, `intent:`)
/// that a server or a link could carry is refused.
const Set<String> _webSchemes = {'https', 'http'};

/// Opens web pages (legal documents) in the platform's in-app browser view
/// (Custom Tabs / SFSafariViewController), so the app stays one tap away.
/// Tests override it with a recorder.
final urlOpenerProvider = Provider<UrlOpener>(
  (ref) => openWebPage,
  name: 'urlOpenerProvider',
);

/// Opens [url] in the in-app browser view when it is a web page.
Future<bool> openWebPage(Uri url) {
  if (!_webSchemes.contains(url.scheme)) return Future.value(false);
  return launchUrl(url, mode: LaunchMode.inAppBrowserView);
}
