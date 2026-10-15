# Guía: tu declaración anual estimada

> **Esto es un aproximado; tu contador te dará las cifras reales.**
> La pantalla "Declaración anual" es una **estimación**. No presenta nada ante el SAT, no
> sustituye la declaración que prepara tu contador y no garantiza el saldo final.

Esta guía es para quien usa la app (persona física o persona moral). Los nombres entre
comillas son los que ves en pantalla. La referencia técnica del módulo está en
[docs/modules/annual.md](../modules/annual.md).

## 1. Dónde está

1. En **Inicio**, toca **"Declaración anual"**, en el encabezado. Si no lo ves, el módulo
   no está activo en tu cuenta.
2. Elige el **"Ejercicio"**. Se puede estimar desde 2026 hasta el año en curso. Del 1 de
   enero al 30 de abril la app abre el año anterior, que es el que se declara.
3. Si el año no ha terminado, verás **"Año en curso"**. La app calcula con tus registros
   hasta el último mes cerrado y no proyecta los meses que faltan.

## 2. Qué ves

- **El saldo:** "Saldo a cargo" (lo que pagarías), "Saldo a favor" (lo que te podrían
  devolver) o "Sin saldo". Debajo aparece la "Fecha límite" y siempre la leyenda **"Esto es
  un aproximado; tu contador te dará las cifras reales."**
- **Una sección por grupo de regímenes**, con su estado:
  - **"Completo"**: la app tiene todo lo que necesita para esa sección.
  - **"Incompleto"**: falta algo. Debajo de **"Falta:"** aparece cada dato con un enlace
    a la pantalla donde se captura. La cifra usa cero en lo que falta, así que puede cambiar
    bastante cuando lo captures.
  - **"Informativo"**: se muestra para que lo sepas, pero no es un saldo. Un ejemplo es
    "RESICO: revisión informativa": tus pagos mensuales de RESICO son definitivos y la
    "Diferencia informativa" no se paga ni se devuelve.
- **"Ver cómo se calculó"**: el desglose paso a paso, cada línea con su fundamento legal y
  su origen ("De tus constancias de sueldos", "Pago presentado", "Estimado por la app"…).
- **"Requiere validación de un contador"**: un aviso destacado cuando el cálculo usa una
  regla que un contador todavía no confirma (por ejemplo, el uso de la UMA en las
  deducciones de dependientes). Revísalo con tu contador antes de confiar en la cifra.
- **Avisos al final**, siempre presentes: que es una estimación, que el IVA no tiene
  declaración anual (se paga mes con mes) y, para 2026, que las facilidades de la anual se
  publican en la RMF 2027 y pueden cambiar.

Fechas límite del ejercicio 2026 (pueden moverse al siguiente día hábil o si el SAT las
prorroga):

| Quién | Fecha límite |
|---|---|
| Personas físicas | 30 de abril de 2027 |
| Personas morales | 31 de marzo de 2027 |
| Personas morales con fines no lucrativos (603), declaración informativa | 15 de febrero de 2027 |

## 3. Qué capturar: "Datos para tu anual"

Debajo del saldo está el bloque **"Datos para tu anual"**. Cada renglón lleva a una
pantalla y dice cuántos datos llevas. Los enlaces "Falta:" de una sección "Incompleto" te
llevan a las mismas pantallas. Al volver, la estimación se actualiza sola.

| Renglón | Quién lo ve | Para qué sirve |
|---|---|---|
| "Deducciones personales" | Personas físicas | Bajan la base del impuesto (§3.1) |
| "Constancias de sueldos" | Personas físicas | Tus sueldos (régimen 605) y el ISR que te retuvieron (§3.2) |
| "Pagos provisionales presentados" | Todos, uno por régimen con pagos mensuales | Lo que de verdad pagaste cada mes (§3.3) |
| "Datos del año" | Todos | Datos que la app no puede deducir de tus registros (§3.4) |

### 3.1 Deducciones personales (personas físicas)

Toca **"Agregar deducción"**. Elige primero el **"Tipo de deducción"**: la app sólo te pide
los datos de ese tipo. Captura el **"Monto"** total de tu factura o constancia y la
**"Fecha de pago"**, que debe caer dentro del ejercicio.

Tipos: "Honorarios médicos y hospitalarios", "Gastos médicos por incapacidad o
discapacidad", "Gastos funerarios", "Donativos", "Intereses reales de crédito
hipotecario", "Aportaciones voluntarias para el retiro", "Primas de seguros de gastos
médicos", "Transporte escolar obligatorio", "Impuesto local sobre salarios" y
"Colegiaturas".

**Topes de ley (ejercicio 2026).** La app aplica los topes por ti. En la lista ves dos
cifras: "Registrado" (lo que capturaste) y "Aplicado en tu estimación" (lo que cuenta
después de los topes). Cada tope que recorta algo aparece como **"Tope aplicado"**.

| Concepto | Tope |
|---|---|
| Tope global (médicos, funerarios, donativos, intereses hipotecarios, seguros, transporte escolar, impuesto local) | El menor de 5 UMA anuales ($213,973.20) o 15 % de tus ingresos totales |
| Gastos funerarios | 1 UMA anual ($42,794.64) |
| Donativos | 7 % de tus ingresos acumulables del año anterior; a gobierno, hasta 4 % |
| Aportaciones voluntarias para el retiro | El menor de 10 % de tus ingresos acumulables o 5 UMA anuales; fuera del tope global |
| Gastos médicos por incapacidad o discapacidad | Sin tope |
| Colegiaturas, por alumno y por el nivel más alto del año | Preescolar $14,200 · Primaria $12,900 · Secundaria $19,900 · Profesional técnico $17,100 · Bachillerato $24,500; fuera del tope global |

Cuidado con:

- **Efectivo.** En honorarios médicos, gastos por incapacidad, transporte escolar y
  colegiaturas, un pago en efectivo **no es deducible**. Marca "Pagué con tarjeta,
  transferencia o cheque" si fue así. Si no, la app guarda la deducción pero la deja fuera
  ("En efectivo: no deducible").
- **Para quién fue.** Si la deducción fue para tu cónyuge, tus padres o tus hijos, confirma
  que esa persona tuvo ingresos menores a una UMA anual en el año. La app no lo verifica:
  confías en tu propia declaración.
- **Colegiaturas.** Captura sólo la colegiatura, sin inscripción ni reinscripción. Escribe
  un nombre en "Alumno" para distinguir a cada hijo: el tope es por alumno. Si recibió beca,
  anótala en "Beca recibida".
- **Donativos.** Necesitan tus ingresos acumulables del año anterior (§3.4). Sin ese dato,
  la app deja los donativos fuera y marca la sección "Incompleto".
- "Folio fiscal (UUID) del CFDI" es opcional. La app no lo consulta con el SAT.

Para borrar una deducción, elimínala y toca **"Deshacer"** si te equivocaste: el aviso no
desaparece solo. Puedes registrar hasta 500 deducciones por año.

### 3.2 Constancias de sueldos (régimen 605)

Si recibes sueldo, captura **una constancia por cada empleador**: el CFDI de retenciones o
la constancia de sueldos que te entrega tu patrón. Toca **"Agregar constancia"** y copia los
totales tal como vienen:

- "Ingresos gravados (acumulables)"
- "Ingresos exentos"
- "ISR retenido"
- "Nombre del empleador" (opcional y sólo para que la reconozcas; la app no pide el RFC
  del empleador)

En **"Situaciones especiales"** marca lo que aplique: "Trabajé para dos o más empleadores al
mismo tiempo", "Dejé de trabajar antes del 31 de diciembre" o "Mi empleador no estaba
obligado a retenerme". Estas casillas deciden si **estás obligado** a presentar la anual.

Mientras tengas el régimen 605 y no hayas capturado ninguna constancia, la sección sale
"Incompleto" con "Falta: Constancias de sueldos" y la app cuenta tu sueldo como cero.
Borrar una constancia **no se puede deshacer**; la app te pide confirmarlo. Puedes
registrar hasta 20 constancias por año.

### 3.3 Pagos provisionales presentados (confirmados)

Por cada régimen con pagos mensuales verás 12 meses. Si un mes no está confirmado, la app lo
rellena con su propia estimación, marcada **"Estimado por la app — confírmalo"** (o "De tus
datos fiscales — confírmalo" si ya capturaste ese pago en los datos fiscales del régimen).
Mientras haya meses sin confirmar, la estimación anual lleva un aviso.

- **Lo mejor:** toma tus acuses de pago del SAT y escribe en cada mes lo que pagaste.
- **"Confirmar todos"** guarda como pagados los montos que ves en los meses sin confirmar.
  Úsalo sólo si coinciden con tus acuses.
- Un mes en el que no tenías el régimen aparece como "No tenías este régimen en este mes."

**¿Por qué un mes dice $0.00 aquí y otra cifra en la estimación mensual?** La pantalla
mensual calcula como si no hubieras pagado los meses anteriores. La anual, en cambio,
supone que sí pagaste lo que la app estimó para esos meses, así que a cada mes le toca sólo
la diferencia. Ninguna de las dos cifras es tu acuse: **escribe lo que de verdad pagaste**.

Un saldo a favor que dependa de pagos "Estimado por la app" lleva la nota "Solo se devuelve
lo efectivamente pagado o retenido". Si el saldo a favor es de $10,000 o más, la app te
recuerda que para pedir la devolución necesitas e.firma.

### 3.4 Datos del año

- **"Ingresos acumulables de {año anterior}"**, antes de deducciones personales. Sirve
  para el tope de donativos. Déjalo vacío si no lo sabes, pero entonces tus donativos no
  cuentan.
- **"Otros ingresos exentos"**: los exentos que no registraste en la app. Cuentan para el
  tope de 15 %.
- **"Le avisé a mi empleador que presentaré mi declaración anual"**: márcalo si lo hiciste.
  Te obliga a presentarla.
- **"Datos de tus regímenes"**: las pérdidas de años anteriores, la PTU pagada y la
  depreciación del año. Se capturan en los "Datos fiscales" de cada régimen; aquí sólo se
  muestran con un enlace para editarlos.

## 4. Personas morales

Una persona moral sólo ve "Pagos provisionales presentados" y "Datos del año": no tiene
deducciones personales ni constancias. Además del saldo, la sección muestra **"Datos
informativos"**: la "Base de PTU", la "PTU a repartir" y el coeficiente de utilidad sugerido
para tus pagos del año siguiente. Confírmalos con tu contador: la tasa de PTU y el redondeo
del coeficiente están marcados como pendientes de validación. La app **no** incluye el
ajuste anual por inflación ni actualiza pérdidas: pídeselos a tu contador.

## 5. Cuenta de sólo consulta

Cuando termina tu periodo de prueba sin un plan, verás "Tu cuenta es de solo consulta". Puedes
ver la estimación y el desglose, y también **eliminar** (y deshacer) tus datos. Capturar o
cambiar algo requiere un plan: al guardar se abre la pantalla de planes y lo que escribiste
no se pierde.

## 6. Lo que la app no hace

- No presenta la declaración ni se conecta con el SAT.
- No calcula IVA anual (no existe) ni el ajuste anual por inflación.
- No recalcula los exentos de tus sueldos ni el subsidio para el empleo: usa los totales de
  tu constancia.
- No aplica facilidades que aún no se publican para el ejercicio (las de la RMF 2027 para
  la anual 2026).
- No garantiza una devolución: sólo se devuelve lo efectivamente pagado o retenido.

Lleva a tu contador la estimación y el desglose ("Ver cómo se calculó"): **él te dará las
cifras reales.**
