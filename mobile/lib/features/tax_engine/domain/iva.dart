import 'package:control_contable/api/generated/models/iva_components.dart';

/// Whether a regime adds anything to the month's consolidated IVA: any of
/// its [components] is non-zero (MODULE-tax-engine §2.6).
bool contributesIva(IvaComponents? components) =>
    components != null &&
    (components.chargedCollectedCents != 0 ||
        components.creditablePaidCents != 0 ||
        components.withheldFromTaxpayerCents != 0 ||
        components.definitiveExcludedCents != 0);
