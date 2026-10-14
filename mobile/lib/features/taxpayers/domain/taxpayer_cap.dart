/// Taxpayers one user may hold, archived ones included (OD-2, OQ-4, OQ-9).
/// UX only: the server's `422 taxpayer_limit_reached` decides.
const int taxpayerCap = 10;

/// Whether a user holding [held] taxpayers can register no more.
bool isAtTaxpayerCap(int held) => held >= taxpayerCap;
