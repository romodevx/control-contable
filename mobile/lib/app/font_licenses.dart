import 'package:flutter/foundation.dart';
import 'package:flutter/services.dart';

/// OFL notice of every face bundled in pubspec.yaml `fonts:`, keyed by family.
///
/// SIL OFL 1.1 requires the notice to travel with each copy of the font. The
/// subset .ttf files do not all carry it in their metadata, so the app ships
/// the text files as assets and lists them in its licenses page.
const _oflNoticeByFamily = {
  'Public Sans': 'assets/fonts/OFL-PublicSans.txt',
  'Source Serif 4': 'assets/fonts/OFL-SourceSerif4.txt',
};

/// Families whose notice [registerFontLicenses] adds.
Iterable<String> get bundledFontFamilies => _oflNoticeByFamily.keys;

/// Adds each bundled face's OFL notice to the [LicenseRegistry], read lazily
/// from the asset bundle when the licenses page asks for it.
void registerFontLicenses() {
  LicenseRegistry.addLicense(() async* {
    for (final MapEntry(key: family, value: asset)
        in _oflNoticeByFamily.entries) {
      final notice = await rootBundle.loadString(asset);
      yield LicenseEntryWithLineBreaks([family], notice);
    }
  });
}
