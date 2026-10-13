# Design directions — Control Contable

| Field | Value |
|---|---|
| Status | DECIDED 2026-10-13 — owner chose **A · Confianza sobria** as-is (no B/C borrowings) |
| Date | 2026-10-13 |
| Preview | [`preview.html`](preview.html) (open it in a browser; each screen is shown in all three directions, in light and dark mode) |
| Inputs | [PRD](../project-specs/PRD.md) §3 personas, §6 screens; `~/.claude/rules/ui-forms-and-modals.md` (house form rules, which win over the style catalog); `ui-ux-pro-max` catalog (Banking/Traditional Finance palette, "Corporate Trust" pairing); `design-consultation` method (face selection by audience world and mode) |

## What every direction shares

These are fixed. Only the look changes between directions.

- **Audience and mode.** Freelancers, landlords, platform workers and small-company owners,
  some of them older, using the app on a phone. Every screen is an *Operate* surface (a task
  to finish) except the breakdown, which is *Read*. Trust and legibility come before personality.
- **Mandatory legend.** "Esto es un aproximado. No todos los gastos aplican para tu giro
  comercial." appears in a tinted block **right under every estimate** (dashboard hero,
  breakdown summary, history month). It is never hidden in a footer and never dismissible.
- **Numbers.** Tabular figures everywhere money appears. Cents are shown smaller on hero
  amounts. Negative lines use a real minus sign (−). Negative IVA is labeled *saldo a favor*
  and uses the `favor` token plus text, never color alone.
- **Forms (house rule).** One column; label above each field; input text 17 px; inputs 52 px
  tall; every target ≥ 44 × 44 px; defaults prefilled (regime, 16 % IVA, today). "Registrar
  ingreso" has 5 visible fields plus 3 optional ones behind "Retenciones y fecha de factura",
  8 in total, so it is **its own page**, not a sheet. Footer: *Guardar* (primary), *Guardar y
  registrar otro* (secondary), *Cancelar* (text). The primary button is never disabled.
  Required-but-empty errors show only on submit; format errors on blur.
- **Navigation.** Bottom bar with five labeled items: Inicio, Ingresos, Gastos, Historial,
  Ajustes. The taxpayer switcher sits at the top of every main screen, showing name, type and
  masked RFC.
- **Charts.** One small chart on Home: six months of estimated tax as stacked bars (IVA at the
  bottom, ISR on top), a visible key, the current month's total as text, and an `aria-label`
  sentence that says the same thing. No pie charts, no gradients. Mark colors are ≥ 3:1
  against the card; text on charts is ≥ 4.5:1.
- **Spacing.** 4-point grid: 4, 8, 12, 16, 20, 24, 32, 48. 16 px screen gutters, 20–24 px
  between sections.
- **Dark mode.** Designed with the light mode, not inverted. Each token has a dark value
  checked separately.
- **Flutter mapping.** Tokens become a `ThemeExtension<AppTokens>` next to a Material 3
  `ColorScheme` (`primary`, `onPrimary`, `surface`, `onSurface`, `outline`, `error`). Fonts
  load through the `google_fonts` package and are bundled as assets for release builds
  (works offline, no runtime fetch). The type scale is set in `TextTheme`, so system text
  scaling still applies; layouts must survive 200 % text.

## Contrast method

WCAG 2.x relative luminance, computed with a script (not estimated by eye). Thresholds:
text ≥ 4.5:1; large text, chart marks and input borders (`outline`) ≥ 3:1 (WCAG 1.4.11).
Decorative dividers (`border`) are exempt and are never the only boundary of a control.
All pairs below pass AA.

---

## A · Confianza sobria (sober trust)

**Idea:** the app reads like a well-kept official document. Institutional ink blue, a serif
for the numbers that matter, small radii, no shadows. It borrows the visual language of a
bank statement or a SAT acknowledgement, without copying either.

### Palette

| Token | Light | Dark | Use |
|---|---|---|---|
| `bg` | `#F5F7FA` | `#0B1420` | Screen background |
| `surface` | `#FFFFFF` | `#13202F` | Cards, inputs, bottom bar |
| `text` | `#0E1B2C` | `#E8EEF5` | Body text |
| `muted` | `#4A5568` | `#A9B6C6` | Secondary text, help |
| `outline` | `#7A8797` | `#6B7D93` | Input and button borders |
| `border` | `#E1E6ED` | `#24344A` | Decorative dividers |
| `primary` | `#1D4E89` | `#8DB8EC` | Primary button, active nav, links |
| `onPrimary` | `#FFFFFF` | `#0B1420` | Text on primary |
| `heroBg` / `heroText` / `heroMuted` | `#1D4E89` / `#FFFFFF` / `#C7D7EC` | `#16365C` / `#E8EEF5` / `#B4C6DD` | Estimate card |
| `iva` (chart) | `#1D4E89` | `#8DB8EC` | IVA series |
| `isr` (chart) | `#2B7470` | `#6FC7C0` | ISR series |
| `favor` | `#1C7046` | `#6FD49A` | Saldo a favor |
| `legendBg` / `legendText` | `#FFF4D6` / `#5C3F00` | `#2B2310` / `#F2D58A` | Mandatory legend |
| `error` | `#B42318` | `#FF8A80` | Errors |

| Pair | Light | Dark |
|---|---|---|
| text / bg | 16.1 | 15.8 |
| text / surface | 17.3 | 14.1 |
| muted / surface | 7.5 | 8.0 |
| onPrimary / primary | 8.4 | 9.0 |
| heroText / heroBg | 8.4 | 10.5 |
| heroMuted / heroBg | 5.7 | 7.0 |
| iva mark / surface | 8.4 | 8.0 |
| isr mark / surface | 5.5 | 8.3 |
| favor / surface | 6.1 | 9.1 |
| legendText / legendBg | 8.8 | 10.9 |
| error / surface | 6.6 | 7.2 |
| outline / surface | 3.7 | 3.9 |

### Typography

| Role | Face | Weights | Why |
|---|---|---|---|
| Display (hero amounts, screen titles, section titles) | **Source Serif 4** | 600, 700 | Serif amounts read as "a document", which is the trust signal. Optical sizes keep 40–44 px figures sharp |
| Body / UI / numbers | **Public Sans** | 400, 500, 600, 700 | Made for US government interfaces: neutral, wide, very legible at 16–17 px, good tabular figures |

Scale (px): 44 hero · 28 page · 20 title · 18 section · 17 input/button · 16 body · 14 help · 13 caption.

- **Spacing / radius:** radii 6 / 8 / 12 (control / button / card). No elevation; cards are
  separated by a 1 px border.
- **Icons:** Lucide (`lucide_icons_flutter`), outline, 1.75 stroke, 22 px.
- **Chart:** square-topped stacked bars, navy IVA under teal ISR. The active nav item gets a
  3 px top rule.
- **Copy tone:** formal and impersonal, avoiding both *tú* and *usted* where possible ("Vence
  el 17 de noviembre", "Registrar ingreso"). Short, factual, with the legal basis on every line.

**Pros**
- The strongest trust signal for older users and for company owners who show the screen to
  an accountant.
- The navy hero makes "how much do I owe" the obvious first read.
- Serif plus sans gives clear hierarchy without extra color.

**Cons**
- Closest to a bank app: it does not stand out on the store.
- The filled navy hero is heavy. It needs a light empty state ("Registra tu primer ingreso")
  so a new user does not see a big blue $0.00.

---

## B · Fintech cálido (warm fintech)

**Idea:** a friendly helper rather than an authority. Clay and cream, round corners, soft
shadows in light mode. The goal is to lower the anxiety that comes with the word "SAT".

### Palette

| Token | Light | Dark | Use |
|---|---|---|---|
| `bg` | `#FBF7F2` | `#1A1411` | Screen background |
| `surface` | `#FFFFFF` | `#251C17` | Cards, inputs |
| `text` | `#2A1F17` | `#F5ECE4` | Body |
| `muted` | `#6B5B4E` | `#C2B2A4` | Secondary |
| `outline` | `#9A8676` | `#8A7768` | Control borders |
| `border` | `#EFE5DA` | `#3A2D25` | Dividers |
| `primary` | `#A13F1C` | `#F0A07E` | Primary, active nav |
| `onPrimary` | `#FFFFFF` | `#1A1411` | Text on primary |
| `heroBg` / `heroText` / `heroMuted` | `#F7E3D3` / `#2A1F17` / `#6B4A38` | `#3A271C` / `#F5ECE4` / `#D8C2B0` | Estimate card |
| `iva` (chart) | `#A13F1C` | `#F0A07E` | IVA series |
| `isr` (chart) | `#3E6B5A` | `#8FC5AE` | ISR series |
| `favor` | `#2A7148` | `#7DD3A0` | Saldo a favor |
| `legendBg` / `legendText` | `#FCEFD9` / `#6E4A12` | `#33281A` / `#F3CF92` | Legend |
| `error` | `#B3261E` | `#FF9A8F` | Errors |

| Pair | Light | Dark |
|---|---|---|
| text / bg | 15.1 | 15.6 |
| text / surface | 16.1 | 14.3 |
| muted / surface | 6.5 | 8.1 |
| onPrimary / primary | 6.5 | 8.7 |
| heroText / heroBg | 12.9 | 12.1 |
| heroMuted / heroBg | 6.3 | 8.2 |
| iva mark / surface | 6.5 | 8.0 |
| isr mark / surface | 6.1 | 8.6 |
| favor / surface | 5.9 | 9.3 |
| legendText / legendBg | 7.0 | 9.7 |
| error / surface | 6.5 | 8.2 |
| outline / surface | 3.5 | 3.9 |

### Typography

| Role | Face | Weights | Why |
|---|---|---|---|
| Display | **Bricolage Grotesque** | 500, 700 | A grotesque with some character at large sizes; gives the brand a voice without being cute |
| Body / UI / numbers | **Figtree** | 400, 500, 600, 700 | Friendly geometric sans with open shapes; reads well at 16 px |

- **Spacing / radius:** radii 12 / 16 / 22. Light mode uses a two-layer warm shadow on
  cards; dark mode drops shadows and relies on surface steps.
- **Icons:** Phosphor Regular (`phosphor_flutter`), 2 px stroke, rounded joins. The active
  nav item gets a pill behind the icon (Material 3 style).
- **Chart:** stacked bars with 5 px rounded tops, clay IVA under sage ISR.
- **Copy tone:** warm *tú*: "Esto es lo que calculamos que pagarás", "¡Listo! Guardamos tu
  ingreso". The legend still uses the exact mandated text.

**Pros**
- The most approachable; good for freelancers and platform workers who are new to their
  obligations.
- Stands out among blue finance apps.
- Rounded, larger shapes are forgiving for touch.

**Cons**
- Clay red sits close to "error / debt" red. Errors must always carry an icon and text, and
  amounts owed must never be colored clay.
- Can read as less serious to company owners (PM 601) and to accountants.
- Shadows and two display weights mean more design upkeep.

---

## C · Minimal claro (clear minimal)

**Idea:** a ledger. Black on white, one green accent, monospaced figures, hairline rules
instead of cards. Nothing on screen except the numbers and how they were reached.

### Palette

| Token | Light | Dark | Use |
|---|---|---|---|
| `bg` | `#FFFFFF` | `#0D0D0D` | Background |
| `surface` | `#F6F6F4` | `#1A1A1A` | Inputs, bottom bar |
| `text` | `#111111` | `#F2F2F2` | Body |
| `muted` | `#555555` | `#A8A8A8` | Secondary |
| `outline` | `#8A8A87` | `#6E6E6E` | Control borders |
| `border` | `#E6E6E3` | `#2A2A2A` | Dividers |
| `primary` | `#111111` | `#F2F2F2` | Primary button |
| `onPrimary` | `#FFFFFF` | `#0D0D0D` | Text on primary |
| `accent` (= `isr`, `favor`, focus) | `#0B7360` | `#4CC9A8` | The only color |
| `iva` (chart) | `#111111` | `#F2F2F2` | IVA series |
| `legendBg` / `legendText` | `#EFEFEC` / `#333333` | `#242424` / `#D6D6D6` | Legend (plus a 3 px accent rule) |
| `error` | `#C62828` | `#FF8A80` | Errors |

| Pair | Light | Dark |
|---|---|---|
| text / bg | 18.9 | 17.4 |
| text / surface | 17.5 | 15.5 |
| muted / surface | 6.9 | 7.3 |
| onPrimary / primary | 18.9 | 17.4 |
| iva mark / surface | 17.5 | 15.5 |
| accent / surface | 5.3 | 8.5 |
| legendText / legendBg | 11.0 | 10.7 |
| error / surface | 5.2 | 7.6 |
| outline / surface | 3.2 | 3.4 |

### Typography

| Role | Face | Weights | Why |
|---|---|---|---|
| Display | **Manrope** | 600, 700, 800 | Clean, slightly technical; uppercase section labels with tracking |
| Body / UI | **Atkinson Hyperlegible** | 400, 700 | Designed by the Braille Institute for low-vision readers; clearly distinct letterforms (Il1, O0) |
| Numbers | **IBM Plex Mono** | 400, 500, 600 | Columns of amounts line up like a ledger (IBM Plex Mono is on the approved mono list for data) |

- **Spacing / radius:** radii 4 / 4 / 6. No cards: sections open with a 1 px rule in `text`
  color; results get a 2 px rule.
- **Icons:** Material Symbols Outlined, weight 300 (built into Flutter via
  `material_symbols_icons`), 1.5 stroke look.
- **Chart:** flat stacked bars, black IVA under green ISR, no rounding.
- **Copy tone:** terse and neutral: "Estimado: $8,412.00", "Vence 17 nov".

**Pros**
- The most legible numbers; the breakdown reads like an accountant's working paper.
- The cheapest to build and keep consistent: almost no color decisions.
- Atkinson Hyperlegible is the best choice here for older and low-vision users.

**Cons**
- It can feel cold or "unfinished" to non-expert users; the hero has no container, so the
  estimate is less obvious at a glance.
- Monospaced amounts take more width; long amounts ($1,234,567.89) wrap sooner at 390 px
  with large text.
- Three families means three font assets to bundle.

---

## Recommendation

**A · Confianza sobria**, with two borrowings:

1. From **C**, the ledger treatment of the breakdown (hairline rules, a 2 px rule above each
   result) and Atkinson Hyperlegible as a candidate body face if testing with older users
   shows Public Sans is not enough.
2. From **B**, the warmer copy in empty and success states ("Registra tu primer ingreso y te
   mostramos cuánto podrías pagar"), so the formal look does not feel cold to first-time
   users.

Why A: the product's job is to make an estimate *believable* (PRD G2 and the "opens the
breakdown monthly" trust metric). The personas include older users and company owners who
will compare the number with their accountant's. A document-like look supports that better
than B's friendliness, and is easier to recognize at a glance than C's bare ledger. Its
contrast margins are also the widest of the three in both modes.

## Next steps once the owner chooses

1. Freeze the chosen tokens as `docs/design/DESIGN.md` (front-matter tokens plus rationale).
2. Implement `AppTokens` + `ThemeData` (light and dark) in `mobile/`, with widget golden tests
   per theme.
3. Bundle the chosen Google Fonts as assets (`google_fonts` with `allowRuntimeFetching = false`
   in release).

## Decision (2026-10-13)

The owner chose **A · Confianza sobria** as presented, without the B/C borrowings suggested
in the recommendation. The mobile theme tokens in `mobile/` (P0-22) already implement A; any
later change of direction is a token swap, not a widget change.
