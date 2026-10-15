// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'income_create.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$IncomeCreate {

@JsonKey(name: 'regime_code') RegimeCode get regimeCode;@JsonKey(name: 'counterparty_type')@Deprecated('This is marked as deprecated') CounterpartyType get counterpartyType;/// Subtotal before IVA, MXN centavos.
@JsonKey(name: 'amount_cents') PositiveCents get amountCents;@JsonKey(name: 'iva_rate') IvaRate get ivaRate;/// IVA charged; `0` for `rate_0`/`exempt`, else within 100 centavos of `amount_cents × rate`.
@JsonKey(name: 'iva_cents') NonNegativeCents get ivaCents;/// Client-generated idempotency key (UUIDv7 recommended), once per form session.
@JsonKey(includeIfNull: false) Uuid? get id;/// Required since 0.10.1 (`400 required` when missing).
@JsonKey(includeIfNull: false, name: 'counterparty_kind') CounterpartyKind? get counterpartyKind;/// Asked only when the payer can withhold (`pm`, `gov_*`, not via a platform); omitted → the last value used in this regime, else the regime default (`IncomeLastUsed`).
@JsonKey(includeIfNull: false, name: 'operation_kind') OperationKind? get operationKind;/// Housing rent (`rent_real_estate` only; default `false`). `true` requires `iva_rate: exempt` and `iva_cents: 0` (LIVA 20 fr. II), else `422` `residential_rent_exempt` on `iva_rate`.
@JsonKey(includeIfNull: false, name: 'residential_rent') bool? get residentialRent;/// Collected through a digital platform (default `true` in 625, `false` otherwise). Allowed for 625 and persona moral regimes; `true` on any other regime answers `400` `not_for_regime`.
@JsonKey(includeIfNull: false, name: 'via_platform') bool? get viaPlatform;/// Retired by WH-06 (0.10.1, ADR-0008). Any value answers `400 additionalProperties`; the server derives it.
@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') NonNegativeCents? get isrWithheldCents;/// Retired by WH-06 (0.10.1, ADR-0008). Any value answers `400 additionalProperties`; the server derives it.
@JsonKey(includeIfNull: false, name: 'iva_withheld_cents') NonNegativeCents? get ivaWithheldCents;@JsonKey(includeIfNull: false, name: 'invoiced_on') BusinessDate? get invoicedOn;@JsonKey(includeIfNull: false, name: 'collected_on') BusinessDate? get collectedOn;@JsonKey(includeIfNull: false, name: 'platform_activity') PlatformActivity? get platformActivity;@JsonKey(includeIfNull: false, name: 'cfdi_uuid') CfdiUuid? get cfdiUuid;@JsonKey(includeIfNull: false) RecordDescription? get description;
/// Create a copy of IncomeCreate
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$IncomeCreateCopyWith<IncomeCreate> get copyWith => _$IncomeCreateCopyWithImpl<IncomeCreate>(this as IncomeCreate, _$identity);

  /// Serializes this IncomeCreate to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as IncomeCreate;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is IncomeCreate&&(identical(other.regimeCode, _this.regimeCode) || other.regimeCode == _this.regimeCode)&&(identical(other.counterpartyType, _this.counterpartyType) || other.counterpartyType == _this.counterpartyType)&&(identical(other.amountCents, _this.amountCents) || other.amountCents == _this.amountCents)&&(identical(other.ivaRate, _this.ivaRate) || other.ivaRate == _this.ivaRate)&&(identical(other.ivaCents, _this.ivaCents) || other.ivaCents == _this.ivaCents)&&(identical(other.id, _this.id) || other.id == _this.id)&&(identical(other.counterpartyKind, _this.counterpartyKind) || other.counterpartyKind == _this.counterpartyKind)&&(identical(other.operationKind, _this.operationKind) || other.operationKind == _this.operationKind)&&(identical(other.residentialRent, _this.residentialRent) || other.residentialRent == _this.residentialRent)&&(identical(other.viaPlatform, _this.viaPlatform) || other.viaPlatform == _this.viaPlatform)&&(identical(other.isrWithheldCents, _this.isrWithheldCents) || other.isrWithheldCents == _this.isrWithheldCents)&&(identical(other.ivaWithheldCents, _this.ivaWithheldCents) || other.ivaWithheldCents == _this.ivaWithheldCents)&&(identical(other.invoicedOn, _this.invoicedOn) || other.invoicedOn == _this.invoicedOn)&&(identical(other.collectedOn, _this.collectedOn) || other.collectedOn == _this.collectedOn)&&(identical(other.platformActivity, _this.platformActivity) || other.platformActivity == _this.platformActivity)&&(identical(other.cfdiUuid, _this.cfdiUuid) || other.cfdiUuid == _this.cfdiUuid)&&(identical(other.description, _this.description) || other.description == _this.description));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as IncomeCreate;
  return Object.hash(runtimeType,_this.regimeCode,_this.counterpartyType,_this.amountCents,_this.ivaRate,_this.ivaCents,_this.id,_this.counterpartyKind,_this.operationKind,_this.residentialRent,_this.viaPlatform,_this.isrWithheldCents,_this.ivaWithheldCents,_this.invoicedOn,_this.collectedOn,_this.platformActivity,_this.cfdiUuid,_this.description);
}

@override
String toString() {
  final _this = this as IncomeCreate;
  return 'IncomeCreate(regimeCode: ${_this.regimeCode}, counterpartyType: ${_this.counterpartyType}, amountCents: ${_this.amountCents}, ivaRate: ${_this.ivaRate}, ivaCents: ${_this.ivaCents}, id: ${_this.id}, counterpartyKind: ${_this.counterpartyKind}, operationKind: ${_this.operationKind}, residentialRent: ${_this.residentialRent}, viaPlatform: ${_this.viaPlatform}, isrWithheldCents: ${_this.isrWithheldCents}, ivaWithheldCents: ${_this.ivaWithheldCents}, invoicedOn: ${_this.invoicedOn}, collectedOn: ${_this.collectedOn}, platformActivity: ${_this.platformActivity}, cfdiUuid: ${_this.cfdiUuid}, description: ${_this.description})';
}


}

/// @nodoc
abstract mixin class $IncomeCreateCopyWith<$Res>  {
  factory $IncomeCreateCopyWith(IncomeCreate value, $Res Function(IncomeCreate) _then) = _$IncomeCreateCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode regimeCode,@JsonKey(name: 'counterparty_type')@Deprecated('This is marked as deprecated') CounterpartyType counterpartyType,@JsonKey(name: 'amount_cents') PositiveCents amountCents,@JsonKey(name: 'iva_rate') IvaRate ivaRate,@JsonKey(name: 'iva_cents') NonNegativeCents ivaCents,@JsonKey(includeIfNull: false) Uuid? id,@JsonKey(includeIfNull: false, name: 'counterparty_kind') CounterpartyKind? counterpartyKind,@JsonKey(includeIfNull: false, name: 'operation_kind') OperationKind? operationKind,@JsonKey(includeIfNull: false, name: 'residential_rent') bool? residentialRent,@JsonKey(includeIfNull: false, name: 'via_platform') bool? viaPlatform,@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') NonNegativeCents? isrWithheldCents,@JsonKey(includeIfNull: false, name: 'iva_withheld_cents') NonNegativeCents? ivaWithheldCents,@JsonKey(includeIfNull: false, name: 'invoiced_on') BusinessDate? invoicedOn,@JsonKey(includeIfNull: false, name: 'collected_on') BusinessDate? collectedOn,@JsonKey(includeIfNull: false, name: 'platform_activity') PlatformActivity? platformActivity,@JsonKey(includeIfNull: false, name: 'cfdi_uuid') CfdiUuid? cfdiUuid,@JsonKey(includeIfNull: false) RecordDescription? description
});




}
/// @nodoc
class _$IncomeCreateCopyWithImpl<$Res>
    implements $IncomeCreateCopyWith<$Res> {
  _$IncomeCreateCopyWithImpl(this._self, this._then);

  final IncomeCreate _self;
  final $Res Function(IncomeCreate) _then;

/// Create a copy of IncomeCreate
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? regimeCode = null,Object? counterpartyType = null,Object? amountCents = null,Object? ivaRate = null,Object? ivaCents = null,Object? id = freezed,Object? counterpartyKind = freezed,Object? operationKind = freezed,Object? residentialRent = freezed,Object? viaPlatform = freezed,Object? isrWithheldCents = freezed,Object? ivaWithheldCents = freezed,Object? invoicedOn = freezed,Object? collectedOn = freezed,Object? platformActivity = freezed,Object? cfdiUuid = freezed,Object? description = freezed,}) {
  return _then(IncomeCreate(
regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,counterpartyType: null == counterpartyType ? _self.counterpartyType : counterpartyType // ignore: cast_nullable_to_non_nullable
as CounterpartyType,amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as PositiveCents,ivaRate: null == ivaRate ? _self.ivaRate : ivaRate // ignore: cast_nullable_to_non_nullable
as IvaRate,ivaCents: null == ivaCents ? _self.ivaCents : ivaCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,id: freezed == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid?,counterpartyKind: freezed == counterpartyKind ? _self.counterpartyKind : counterpartyKind // ignore: cast_nullable_to_non_nullable
as CounterpartyKind?,operationKind: freezed == operationKind ? _self.operationKind : operationKind // ignore: cast_nullable_to_non_nullable
as OperationKind?,residentialRent: freezed == residentialRent ? _self.residentialRent : residentialRent // ignore: cast_nullable_to_non_nullable
as bool?,viaPlatform: freezed == viaPlatform ? _self.viaPlatform : viaPlatform // ignore: cast_nullable_to_non_nullable
as bool?,isrWithheldCents: freezed == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,ivaWithheldCents: freezed == ivaWithheldCents ? _self.ivaWithheldCents : ivaWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,invoicedOn: freezed == invoicedOn ? _self.invoicedOn : invoicedOn // ignore: cast_nullable_to_non_nullable
as BusinessDate?,collectedOn: freezed == collectedOn ? _self.collectedOn : collectedOn // ignore: cast_nullable_to_non_nullable
as BusinessDate?,platformActivity: freezed == platformActivity ? _self.platformActivity : platformActivity // ignore: cast_nullable_to_non_nullable
as PlatformActivity?,cfdiUuid: freezed == cfdiUuid ? _self.cfdiUuid : cfdiUuid // ignore: cast_nullable_to_non_nullable
as CfdiUuid?,description: freezed == description ? _self.description : description // ignore: cast_nullable_to_non_nullable
as RecordDescription?,
  ));
}

}


/// Adds pattern-matching-related methods to [IncomeCreate].
extension IncomeCreatePatterns on IncomeCreate {
/// A variant of `map` that fallback to returning `orElse`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _IncomeCreate value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _IncomeCreate() when $default != null:
return $default(_that);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// Callbacks receives the raw object, upcasted.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case final Subclass2 value:
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _IncomeCreate value)  $default,){
final _that = this;
switch (_that) {
case _IncomeCreate():
return $default(_that);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `map` that fallback to returning `null`.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case final Subclass value:
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _IncomeCreate value)?  $default,){
final _that = this;
switch (_that) {
case _IncomeCreate() when $default != null:
return $default(_that);case _:
  return null;

}
}
/// A variant of `when` that fallback to an `orElse` callback.
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return orElse();
/// }
/// ```

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'counterparty_type')@Deprecated('This is marked as deprecated')  CounterpartyType counterpartyType, @JsonKey(name: 'amount_cents')  PositiveCents amountCents, @JsonKey(name: 'iva_rate')  IvaRate ivaRate, @JsonKey(name: 'iva_cents')  NonNegativeCents ivaCents, @JsonKey(includeIfNull: false)  Uuid? id, @JsonKey(includeIfNull: false, name: 'counterparty_kind')  CounterpartyKind? counterpartyKind, @JsonKey(includeIfNull: false, name: 'operation_kind')  OperationKind? operationKind, @JsonKey(includeIfNull: false, name: 'residential_rent')  bool? residentialRent, @JsonKey(includeIfNull: false, name: 'via_platform')  bool? viaPlatform, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(includeIfNull: false, name: 'iva_withheld_cents')  NonNegativeCents? ivaWithheldCents, @JsonKey(includeIfNull: false, name: 'invoiced_on')  BusinessDate? invoicedOn, @JsonKey(includeIfNull: false, name: 'collected_on')  BusinessDate? collectedOn, @JsonKey(includeIfNull: false, name: 'platform_activity')  PlatformActivity? platformActivity, @JsonKey(includeIfNull: false, name: 'cfdi_uuid')  CfdiUuid? cfdiUuid, @JsonKey(includeIfNull: false)  RecordDescription? description)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _IncomeCreate() when $default != null:
return $default(_that.regimeCode,_that.counterpartyType,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.id,_that.counterpartyKind,_that.operationKind,_that.residentialRent,_that.viaPlatform,_that.isrWithheldCents,_that.ivaWithheldCents,_that.invoicedOn,_that.collectedOn,_that.platformActivity,_that.cfdiUuid,_that.description);case _:
  return orElse();

}
}
/// A `switch`-like method, using callbacks.
///
/// As opposed to `map`, this offers destructuring.
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case Subclass2(:final field2):
///     return ...;
/// }
/// ```

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'counterparty_type')@Deprecated('This is marked as deprecated')  CounterpartyType counterpartyType, @JsonKey(name: 'amount_cents')  PositiveCents amountCents, @JsonKey(name: 'iva_rate')  IvaRate ivaRate, @JsonKey(name: 'iva_cents')  NonNegativeCents ivaCents, @JsonKey(includeIfNull: false)  Uuid? id, @JsonKey(includeIfNull: false, name: 'counterparty_kind')  CounterpartyKind? counterpartyKind, @JsonKey(includeIfNull: false, name: 'operation_kind')  OperationKind? operationKind, @JsonKey(includeIfNull: false, name: 'residential_rent')  bool? residentialRent, @JsonKey(includeIfNull: false, name: 'via_platform')  bool? viaPlatform, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(includeIfNull: false, name: 'iva_withheld_cents')  NonNegativeCents? ivaWithheldCents, @JsonKey(includeIfNull: false, name: 'invoiced_on')  BusinessDate? invoicedOn, @JsonKey(includeIfNull: false, name: 'collected_on')  BusinessDate? collectedOn, @JsonKey(includeIfNull: false, name: 'platform_activity')  PlatformActivity? platformActivity, @JsonKey(includeIfNull: false, name: 'cfdi_uuid')  CfdiUuid? cfdiUuid, @JsonKey(includeIfNull: false)  RecordDescription? description)  $default,) {final _that = this;
switch (_that) {
case _IncomeCreate():
return $default(_that.regimeCode,_that.counterpartyType,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.id,_that.counterpartyKind,_that.operationKind,_that.residentialRent,_that.viaPlatform,_that.isrWithheldCents,_that.ivaWithheldCents,_that.invoicedOn,_that.collectedOn,_that.platformActivity,_that.cfdiUuid,_that.description);case _:
  throw StateError('Unexpected subclass');

}
}
/// A variant of `when` that fallback to returning `null`
///
/// It is equivalent to doing:
/// ```dart
/// switch (sealedClass) {
///   case Subclass(:final field):
///     return ...;
///   case _:
///     return null;
/// }
/// ```

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'counterparty_type')@Deprecated('This is marked as deprecated')  CounterpartyType counterpartyType, @JsonKey(name: 'amount_cents')  PositiveCents amountCents, @JsonKey(name: 'iva_rate')  IvaRate ivaRate, @JsonKey(name: 'iva_cents')  NonNegativeCents ivaCents, @JsonKey(includeIfNull: false)  Uuid? id, @JsonKey(includeIfNull: false, name: 'counterparty_kind')  CounterpartyKind? counterpartyKind, @JsonKey(includeIfNull: false, name: 'operation_kind')  OperationKind? operationKind, @JsonKey(includeIfNull: false, name: 'residential_rent')  bool? residentialRent, @JsonKey(includeIfNull: false, name: 'via_platform')  bool? viaPlatform, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(includeIfNull: false, name: 'iva_withheld_cents')  NonNegativeCents? ivaWithheldCents, @JsonKey(includeIfNull: false, name: 'invoiced_on')  BusinessDate? invoicedOn, @JsonKey(includeIfNull: false, name: 'collected_on')  BusinessDate? collectedOn, @JsonKey(includeIfNull: false, name: 'platform_activity')  PlatformActivity? platformActivity, @JsonKey(includeIfNull: false, name: 'cfdi_uuid')  CfdiUuid? cfdiUuid, @JsonKey(includeIfNull: false)  RecordDescription? description)?  $default,) {final _that = this;
switch (_that) {
case _IncomeCreate() when $default != null:
return $default(_that.regimeCode,_that.counterpartyType,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.id,_that.counterpartyKind,_that.operationKind,_that.residentialRent,_that.viaPlatform,_that.isrWithheldCents,_that.ivaWithheldCents,_that.invoicedOn,_that.collectedOn,_that.platformActivity,_that.cfdiUuid,_that.description);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _IncomeCreate implements IncomeCreate {
  const _IncomeCreate({@JsonKey(name: 'regime_code') required this.regimeCode, @JsonKey(name: 'counterparty_type')@Deprecated('This is marked as deprecated') required this.counterpartyType, @JsonKey(name: 'amount_cents') required this.amountCents, @JsonKey(name: 'iva_rate') required this.ivaRate, @JsonKey(name: 'iva_cents') required this.ivaCents, @JsonKey(includeIfNull: false) this.id, @JsonKey(includeIfNull: false, name: 'counterparty_kind') this.counterpartyKind, @JsonKey(includeIfNull: false, name: 'operation_kind') this.operationKind, @JsonKey(includeIfNull: false, name: 'residential_rent') this.residentialRent, @JsonKey(includeIfNull: false, name: 'via_platform') this.viaPlatform, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents') this.isrWithheldCents, @JsonKey(includeIfNull: false, name: 'iva_withheld_cents') this.ivaWithheldCents, @JsonKey(includeIfNull: false, name: 'invoiced_on') this.invoicedOn, @JsonKey(includeIfNull: false, name: 'collected_on') this.collectedOn, @JsonKey(includeIfNull: false, name: 'platform_activity') this.platformActivity, @JsonKey(includeIfNull: false, name: 'cfdi_uuid') this.cfdiUuid, @JsonKey(includeIfNull: false) this.description});
  factory _IncomeCreate.fromJson(Map<String, dynamic> json) => _$IncomeCreateFromJson(json);

@override@JsonKey(name: 'regime_code') final  RegimeCode regimeCode;
@override@JsonKey(name: 'counterparty_type')@Deprecated('This is marked as deprecated') final  CounterpartyType counterpartyType;
/// Subtotal before IVA, MXN centavos.
@override@JsonKey(name: 'amount_cents') final  PositiveCents amountCents;
@override@JsonKey(name: 'iva_rate') final  IvaRate ivaRate;
/// IVA charged; `0` for `rate_0`/`exempt`, else within 100 centavos of `amount_cents × rate`.
@override@JsonKey(name: 'iva_cents') final  NonNegativeCents ivaCents;
/// Client-generated idempotency key (UUIDv7 recommended), once per form session.
@override@JsonKey(includeIfNull: false) final  Uuid? id;
/// Required since 0.10.1 (`400 required` when missing).
@override@JsonKey(includeIfNull: false, name: 'counterparty_kind') final  CounterpartyKind? counterpartyKind;
/// Asked only when the payer can withhold (`pm`, `gov_*`, not via a platform); omitted → the last value used in this regime, else the regime default (`IncomeLastUsed`).
@override@JsonKey(includeIfNull: false, name: 'operation_kind') final  OperationKind? operationKind;
/// Housing rent (`rent_real_estate` only; default `false`). `true` requires `iva_rate: exempt` and `iva_cents: 0` (LIVA 20 fr. II), else `422` `residential_rent_exempt` on `iva_rate`.
@override@JsonKey(includeIfNull: false, name: 'residential_rent') final  bool? residentialRent;
/// Collected through a digital platform (default `true` in 625, `false` otherwise). Allowed for 625 and persona moral regimes; `true` on any other regime answers `400` `not_for_regime`.
@override@JsonKey(includeIfNull: false, name: 'via_platform') final  bool? viaPlatform;
/// Retired by WH-06 (0.10.1, ADR-0008). Any value answers `400 additionalProperties`; the server derives it.
@override@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') final  NonNegativeCents? isrWithheldCents;
/// Retired by WH-06 (0.10.1, ADR-0008). Any value answers `400 additionalProperties`; the server derives it.
@override@JsonKey(includeIfNull: false, name: 'iva_withheld_cents') final  NonNegativeCents? ivaWithheldCents;
@override@JsonKey(includeIfNull: false, name: 'invoiced_on') final  BusinessDate? invoicedOn;
@override@JsonKey(includeIfNull: false, name: 'collected_on') final  BusinessDate? collectedOn;
@override@JsonKey(includeIfNull: false, name: 'platform_activity') final  PlatformActivity? platformActivity;
@override@JsonKey(includeIfNull: false, name: 'cfdi_uuid') final  CfdiUuid? cfdiUuid;
@override@JsonKey(includeIfNull: false) final  RecordDescription? description;

/// Create a copy of IncomeCreate
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$IncomeCreateCopyWith<_IncomeCreate> get copyWith => __$IncomeCreateCopyWithImpl<_IncomeCreate>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$IncomeCreateToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _IncomeCreate&&(identical(other.regimeCode, regimeCode) || other.regimeCode == regimeCode)&&(identical(other.counterpartyType, counterpartyType) || other.counterpartyType == counterpartyType)&&(identical(other.amountCents, amountCents) || other.amountCents == amountCents)&&(identical(other.ivaRate, ivaRate) || other.ivaRate == ivaRate)&&(identical(other.ivaCents, ivaCents) || other.ivaCents == ivaCents)&&(identical(other.id, id) || other.id == id)&&(identical(other.counterpartyKind, counterpartyKind) || other.counterpartyKind == counterpartyKind)&&(identical(other.operationKind, operationKind) || other.operationKind == operationKind)&&(identical(other.residentialRent, residentialRent) || other.residentialRent == residentialRent)&&(identical(other.viaPlatform, viaPlatform) || other.viaPlatform == viaPlatform)&&(identical(other.isrWithheldCents, isrWithheldCents) || other.isrWithheldCents == isrWithheldCents)&&(identical(other.ivaWithheldCents, ivaWithheldCents) || other.ivaWithheldCents == ivaWithheldCents)&&(identical(other.invoicedOn, invoicedOn) || other.invoicedOn == invoicedOn)&&(identical(other.collectedOn, collectedOn) || other.collectedOn == collectedOn)&&(identical(other.platformActivity, platformActivity) || other.platformActivity == platformActivity)&&(identical(other.cfdiUuid, cfdiUuid) || other.cfdiUuid == cfdiUuid)&&(identical(other.description, description) || other.description == description));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,regimeCode,counterpartyType,amountCents,ivaRate,ivaCents,id,counterpartyKind,operationKind,residentialRent,viaPlatform,isrWithheldCents,ivaWithheldCents,invoicedOn,collectedOn,platformActivity,cfdiUuid,description);
}

@override
String toString() {
    return 'IncomeCreate(regimeCode: $regimeCode, counterpartyType: $counterpartyType, amountCents: $amountCents, ivaRate: $ivaRate, ivaCents: $ivaCents, id: $id, counterpartyKind: $counterpartyKind, operationKind: $operationKind, residentialRent: $residentialRent, viaPlatform: $viaPlatform, isrWithheldCents: $isrWithheldCents, ivaWithheldCents: $ivaWithheldCents, invoicedOn: $invoicedOn, collectedOn: $collectedOn, platformActivity: $platformActivity, cfdiUuid: $cfdiUuid, description: $description)';
}


}

/// @nodoc
abstract mixin class _$IncomeCreateCopyWith<$Res> implements $IncomeCreateCopyWith<$Res> {
  factory _$IncomeCreateCopyWith(_IncomeCreate value, $Res Function(_IncomeCreate) _then) = __$IncomeCreateCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode regimeCode,@JsonKey(name: 'counterparty_type')@Deprecated('This is marked as deprecated') CounterpartyType counterpartyType,@JsonKey(name: 'amount_cents') PositiveCents amountCents,@JsonKey(name: 'iva_rate') IvaRate ivaRate,@JsonKey(name: 'iva_cents') NonNegativeCents ivaCents,@JsonKey(includeIfNull: false) Uuid? id,@JsonKey(includeIfNull: false, name: 'counterparty_kind') CounterpartyKind? counterpartyKind,@JsonKey(includeIfNull: false, name: 'operation_kind') OperationKind? operationKind,@JsonKey(includeIfNull: false, name: 'residential_rent') bool? residentialRent,@JsonKey(includeIfNull: false, name: 'via_platform') bool? viaPlatform,@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') NonNegativeCents? isrWithheldCents,@JsonKey(includeIfNull: false, name: 'iva_withheld_cents') NonNegativeCents? ivaWithheldCents,@JsonKey(includeIfNull: false, name: 'invoiced_on') BusinessDate? invoicedOn,@JsonKey(includeIfNull: false, name: 'collected_on') BusinessDate? collectedOn,@JsonKey(includeIfNull: false, name: 'platform_activity') PlatformActivity? platformActivity,@JsonKey(includeIfNull: false, name: 'cfdi_uuid') CfdiUuid? cfdiUuid,@JsonKey(includeIfNull: false) RecordDescription? description
});




}
/// @nodoc
class __$IncomeCreateCopyWithImpl<$Res>
    implements _$IncomeCreateCopyWith<$Res> {
  __$IncomeCreateCopyWithImpl(this._self, this._then);

  final _IncomeCreate _self;
  final $Res Function(_IncomeCreate) _then;

/// Create a copy of IncomeCreate
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? regimeCode = null,Object? counterpartyType = null,Object? amountCents = null,Object? ivaRate = null,Object? ivaCents = null,Object? id = freezed,Object? counterpartyKind = freezed,Object? operationKind = freezed,Object? residentialRent = freezed,Object? viaPlatform = freezed,Object? isrWithheldCents = freezed,Object? ivaWithheldCents = freezed,Object? invoicedOn = freezed,Object? collectedOn = freezed,Object? platformActivity = freezed,Object? cfdiUuid = freezed,Object? description = freezed,}) {
  return _then(_IncomeCreate(
regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,counterpartyType: null == counterpartyType ? _self.counterpartyType : counterpartyType // ignore: cast_nullable_to_non_nullable
as CounterpartyType,amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as PositiveCents,ivaRate: null == ivaRate ? _self.ivaRate : ivaRate // ignore: cast_nullable_to_non_nullable
as IvaRate,ivaCents: null == ivaCents ? _self.ivaCents : ivaCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,id: freezed == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid?,counterpartyKind: freezed == counterpartyKind ? _self.counterpartyKind : counterpartyKind // ignore: cast_nullable_to_non_nullable
as CounterpartyKind?,operationKind: freezed == operationKind ? _self.operationKind : operationKind // ignore: cast_nullable_to_non_nullable
as OperationKind?,residentialRent: freezed == residentialRent ? _self.residentialRent : residentialRent // ignore: cast_nullable_to_non_nullable
as bool?,viaPlatform: freezed == viaPlatform ? _self.viaPlatform : viaPlatform // ignore: cast_nullable_to_non_nullable
as bool?,isrWithheldCents: freezed == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,ivaWithheldCents: freezed == ivaWithheldCents ? _self.ivaWithheldCents : ivaWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,invoicedOn: freezed == invoicedOn ? _self.invoicedOn : invoicedOn // ignore: cast_nullable_to_non_nullable
as BusinessDate?,collectedOn: freezed == collectedOn ? _self.collectedOn : collectedOn // ignore: cast_nullable_to_non_nullable
as BusinessDate?,platformActivity: freezed == platformActivity ? _self.platformActivity : platformActivity // ignore: cast_nullable_to_non_nullable
as PlatformActivity?,cfdiUuid: freezed == cfdiUuid ? _self.cfdiUuid : cfdiUuid // ignore: cast_nullable_to_non_nullable
as CfdiUuid?,description: freezed == description ? _self.description : description // ignore: cast_nullable_to_non_nullable
as RecordDescription?,
  ));
}


}

// dart format on
