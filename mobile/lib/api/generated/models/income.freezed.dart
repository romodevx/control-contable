// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'income.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$Income {

 Uuid get id;@JsonKey(name: 'taxpayer_id') Uuid get taxpayerId;@JsonKey(name: 'regime_code') RegimeCode get regimeCode;/// Who was billed.
@JsonKey(name: 'counterparty_kind') CounterpartyKind get counterpartyKind;/// As sent, or the default stored when the question was hidden (last used in the regime, else the regime default).
@JsonKey(name: 'operation_kind') OperationKind get operationKind;@JsonKey(name: 'residential_rent') bool get residentialRent;@JsonKey(name: 'via_platform') bool get viaPlatform;@JsonKey(name: 'amount_cents') PositiveCents get amountCents;@JsonKey(name: 'iva_rate') IvaRate get ivaRate;@JsonKey(name: 'iva_cents') NonNegativeCents get ivaCents;/// Read-only mirror of `withholding.isr_cents` (kept for list totals and the estimate).
@JsonKey(name: 'isr_withheld_cents') NonNegativeCents get isrWithheldCents;/// Read-only mirror of `withholding.iva_cents` (kept for list totals and the estimate).
@JsonKey(name: 'iva_withheld_cents') NonNegativeCents get ivaWithheldCents; Withholding get withholding;@JsonKey(includeIfNull: true, name: 'invoiced_on') BusinessDate? get invoicedOn;@JsonKey(includeIfNull: true, name: 'collected_on') BusinessDate? get collectedOn;@JsonKey(includeIfNull: true, name: 'platform_activity') PlatformActivity? get platformActivity;@JsonKey(includeIfNull: true, name: 'cfdi_uuid') CfdiUuid? get cfdiUuid; RecordDescription? get description;/// Accrual income with `collected_on` still `null` ("Por cobrar"); counts for IVA only once collected.
@JsonKey(name: 'pending_collection') bool get pendingCollection;/// Non-blocking notices computed on writes (create, update, restore), never stored; `[]` on reads. Codes: `duplicate_cfdi_uuid` (`cfdi_uuid`), `provisional_parameters` (`collected_on` / `invoiced_on`), `withholding_cpa_unverified` (`withholding`), `foreign_counterparty_not_supported` (`counterparty_kind`). `withholding_differs_from_rate` is retired: never sent since 0.11.3 (no typed amount is left to compare).
 List<Warning> get warnings;@JsonKey(name: 'created_at') DateTime get createdAt;@JsonKey(name: 'updated_at') DateTime get updatedAt;
/// Create a copy of Income
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$IncomeCopyWith<Income> get copyWith => _$IncomeCopyWithImpl<Income>(this as Income, _$identity);

  /// Serializes this Income to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as Income;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is Income&&(identical(other.id, _this.id) || other.id == _this.id)&&(identical(other.taxpayerId, _this.taxpayerId) || other.taxpayerId == _this.taxpayerId)&&(identical(other.regimeCode, _this.regimeCode) || other.regimeCode == _this.regimeCode)&&(identical(other.counterpartyKind, _this.counterpartyKind) || other.counterpartyKind == _this.counterpartyKind)&&(identical(other.operationKind, _this.operationKind) || other.operationKind == _this.operationKind)&&(identical(other.residentialRent, _this.residentialRent) || other.residentialRent == _this.residentialRent)&&(identical(other.viaPlatform, _this.viaPlatform) || other.viaPlatform == _this.viaPlatform)&&(identical(other.amountCents, _this.amountCents) || other.amountCents == _this.amountCents)&&(identical(other.ivaRate, _this.ivaRate) || other.ivaRate == _this.ivaRate)&&(identical(other.ivaCents, _this.ivaCents) || other.ivaCents == _this.ivaCents)&&(identical(other.isrWithheldCents, _this.isrWithheldCents) || other.isrWithheldCents == _this.isrWithheldCents)&&(identical(other.ivaWithheldCents, _this.ivaWithheldCents) || other.ivaWithheldCents == _this.ivaWithheldCents)&&(identical(other.withholding, _this.withholding) || other.withholding == _this.withholding)&&(identical(other.invoicedOn, _this.invoicedOn) || other.invoicedOn == _this.invoicedOn)&&(identical(other.collectedOn, _this.collectedOn) || other.collectedOn == _this.collectedOn)&&(identical(other.platformActivity, _this.platformActivity) || other.platformActivity == _this.platformActivity)&&(identical(other.cfdiUuid, _this.cfdiUuid) || other.cfdiUuid == _this.cfdiUuid)&&(identical(other.description, _this.description) || other.description == _this.description)&&(identical(other.pendingCollection, _this.pendingCollection) || other.pendingCollection == _this.pendingCollection)&&const DeepCollectionEquality().equals(other.warnings, _this.warnings)&&(identical(other.createdAt, _this.createdAt) || other.createdAt == _this.createdAt)&&(identical(other.updatedAt, _this.updatedAt) || other.updatedAt == _this.updatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as Income;
  return Object.hashAll([runtimeType,_this.id,_this.taxpayerId,_this.regimeCode,_this.counterpartyKind,_this.operationKind,_this.residentialRent,_this.viaPlatform,_this.amountCents,_this.ivaRate,_this.ivaCents,_this.isrWithheldCents,_this.ivaWithheldCents,_this.withholding,_this.invoicedOn,_this.collectedOn,_this.platformActivity,_this.cfdiUuid,_this.description,_this.pendingCollection,const DeepCollectionEquality().hash(_this.warnings),_this.createdAt,_this.updatedAt]);
}

@override
String toString() {
  final _this = this as Income;
  return 'Income(id: ${_this.id}, taxpayerId: ${_this.taxpayerId}, regimeCode: ${_this.regimeCode}, counterpartyKind: ${_this.counterpartyKind}, operationKind: ${_this.operationKind}, residentialRent: ${_this.residentialRent}, viaPlatform: ${_this.viaPlatform}, amountCents: ${_this.amountCents}, ivaRate: ${_this.ivaRate}, ivaCents: ${_this.ivaCents}, isrWithheldCents: ${_this.isrWithheldCents}, ivaWithheldCents: ${_this.ivaWithheldCents}, withholding: ${_this.withholding}, invoicedOn: ${_this.invoicedOn}, collectedOn: ${_this.collectedOn}, platformActivity: ${_this.platformActivity}, cfdiUuid: ${_this.cfdiUuid}, description: ${_this.description}, pendingCollection: ${_this.pendingCollection}, warnings: ${_this.warnings}, createdAt: ${_this.createdAt}, updatedAt: ${_this.updatedAt})';
}


}

/// @nodoc
abstract mixin class $IncomeCopyWith<$Res>  {
  factory $IncomeCopyWith(Income value, $Res Function(Income) _then) = _$IncomeCopyWithImpl;
@useResult
$Res call({
 Uuid id,@JsonKey(name: 'taxpayer_id') Uuid taxpayerId,@JsonKey(name: 'regime_code') RegimeCode regimeCode,@JsonKey(name: 'counterparty_kind') CounterpartyKind counterpartyKind,@JsonKey(name: 'operation_kind') OperationKind operationKind,@JsonKey(name: 'residential_rent') bool residentialRent,@JsonKey(name: 'via_platform') bool viaPlatform,@JsonKey(name: 'amount_cents') PositiveCents amountCents,@JsonKey(name: 'iva_rate') IvaRate ivaRate,@JsonKey(name: 'iva_cents') NonNegativeCents ivaCents,@JsonKey(name: 'isr_withheld_cents') NonNegativeCents isrWithheldCents,@JsonKey(name: 'iva_withheld_cents') NonNegativeCents ivaWithheldCents, Withholding withholding,@JsonKey(includeIfNull: true, name: 'invoiced_on') BusinessDate? invoicedOn,@JsonKey(includeIfNull: true, name: 'collected_on') BusinessDate? collectedOn,@JsonKey(includeIfNull: true, name: 'platform_activity') PlatformActivity? platformActivity,@JsonKey(includeIfNull: true, name: 'cfdi_uuid') CfdiUuid? cfdiUuid, RecordDescription? description,@JsonKey(name: 'pending_collection') bool pendingCollection, List<Warning> warnings,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'updated_at') DateTime updatedAt
});


$WithholdingCopyWith<$Res> get withholding;

}
/// @nodoc
class _$IncomeCopyWithImpl<$Res>
    implements $IncomeCopyWith<$Res> {
  _$IncomeCopyWithImpl(this._self, this._then);

  final Income _self;
  final $Res Function(Income) _then;

/// Create a copy of Income
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? taxpayerId = null,Object? regimeCode = null,Object? counterpartyKind = null,Object? operationKind = null,Object? residentialRent = null,Object? viaPlatform = null,Object? amountCents = null,Object? ivaRate = null,Object? ivaCents = null,Object? isrWithheldCents = null,Object? ivaWithheldCents = null,Object? withholding = null,Object? invoicedOn = freezed,Object? collectedOn = freezed,Object? platformActivity = freezed,Object? cfdiUuid = freezed,Object? description = freezed,Object? pendingCollection = null,Object? warnings = null,Object? createdAt = null,Object? updatedAt = null,}) {
  return _then(Income(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid,taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,counterpartyKind: null == counterpartyKind ? _self.counterpartyKind : counterpartyKind // ignore: cast_nullable_to_non_nullable
as CounterpartyKind,operationKind: null == operationKind ? _self.operationKind : operationKind // ignore: cast_nullable_to_non_nullable
as OperationKind,residentialRent: null == residentialRent ? _self.residentialRent : residentialRent // ignore: cast_nullable_to_non_nullable
as bool,viaPlatform: null == viaPlatform ? _self.viaPlatform : viaPlatform // ignore: cast_nullable_to_non_nullable
as bool,amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as PositiveCents,ivaRate: null == ivaRate ? _self.ivaRate : ivaRate // ignore: cast_nullable_to_non_nullable
as IvaRate,ivaCents: null == ivaCents ? _self.ivaCents : ivaCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,isrWithheldCents: null == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,ivaWithheldCents: null == ivaWithheldCents ? _self.ivaWithheldCents : ivaWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,withholding: null == withholding ? _self.withholding : withholding // ignore: cast_nullable_to_non_nullable
as Withholding,invoicedOn: freezed == invoicedOn ? _self.invoicedOn : invoicedOn // ignore: cast_nullable_to_non_nullable
as BusinessDate?,collectedOn: freezed == collectedOn ? _self.collectedOn : collectedOn // ignore: cast_nullable_to_non_nullable
as BusinessDate?,platformActivity: freezed == platformActivity ? _self.platformActivity : platformActivity // ignore: cast_nullable_to_non_nullable
as PlatformActivity?,cfdiUuid: freezed == cfdiUuid ? _self.cfdiUuid : cfdiUuid // ignore: cast_nullable_to_non_nullable
as CfdiUuid?,description: freezed == description ? _self.description : description // ignore: cast_nullable_to_non_nullable
as RecordDescription?,pendingCollection: null == pendingCollection ? _self.pendingCollection : pendingCollection // ignore: cast_nullable_to_non_nullable
as bool,warnings: null == warnings ? _self.warnings : warnings // ignore: cast_nullable_to_non_nullable
as List<Warning>,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,updatedAt: null == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime,
  ));
}
/// Create a copy of Income
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$WithholdingCopyWith<$Res> get withholding {
  
  return $WithholdingCopyWith<$Res>(_self.withholding, (value) {
    return _then(_self.copyWith(withholding: value));
  });
}
}


/// Adds pattern-matching-related methods to [Income].
extension IncomePatterns on Income {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _Income value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _Income() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _Income value)  $default,){
final _that = this;
switch (_that) {
case _Income():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _Income value)?  $default,){
final _that = this;
switch (_that) {
case _Income() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( Uuid id, @JsonKey(name: 'taxpayer_id')  Uuid taxpayerId, @JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'counterparty_kind')  CounterpartyKind counterpartyKind, @JsonKey(name: 'operation_kind')  OperationKind operationKind, @JsonKey(name: 'residential_rent')  bool residentialRent, @JsonKey(name: 'via_platform')  bool viaPlatform, @JsonKey(name: 'amount_cents')  PositiveCents amountCents, @JsonKey(name: 'iva_rate')  IvaRate ivaRate, @JsonKey(name: 'iva_cents')  NonNegativeCents ivaCents, @JsonKey(name: 'isr_withheld_cents')  NonNegativeCents isrWithheldCents, @JsonKey(name: 'iva_withheld_cents')  NonNegativeCents ivaWithheldCents,  Withholding withholding, @JsonKey(includeIfNull: true, name: 'invoiced_on')  BusinessDate? invoicedOn, @JsonKey(includeIfNull: true, name: 'collected_on')  BusinessDate? collectedOn, @JsonKey(includeIfNull: true, name: 'platform_activity')  PlatformActivity? platformActivity, @JsonKey(includeIfNull: true, name: 'cfdi_uuid')  CfdiUuid? cfdiUuid,  RecordDescription? description, @JsonKey(name: 'pending_collection')  bool pendingCollection,  List<Warning> warnings, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'updated_at')  DateTime updatedAt)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _Income() when $default != null:
return $default(_that.id,_that.taxpayerId,_that.regimeCode,_that.counterpartyKind,_that.operationKind,_that.residentialRent,_that.viaPlatform,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.isrWithheldCents,_that.ivaWithheldCents,_that.withholding,_that.invoicedOn,_that.collectedOn,_that.platformActivity,_that.cfdiUuid,_that.description,_that.pendingCollection,_that.warnings,_that.createdAt,_that.updatedAt);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( Uuid id, @JsonKey(name: 'taxpayer_id')  Uuid taxpayerId, @JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'counterparty_kind')  CounterpartyKind counterpartyKind, @JsonKey(name: 'operation_kind')  OperationKind operationKind, @JsonKey(name: 'residential_rent')  bool residentialRent, @JsonKey(name: 'via_platform')  bool viaPlatform, @JsonKey(name: 'amount_cents')  PositiveCents amountCents, @JsonKey(name: 'iva_rate')  IvaRate ivaRate, @JsonKey(name: 'iva_cents')  NonNegativeCents ivaCents, @JsonKey(name: 'isr_withheld_cents')  NonNegativeCents isrWithheldCents, @JsonKey(name: 'iva_withheld_cents')  NonNegativeCents ivaWithheldCents,  Withholding withholding, @JsonKey(includeIfNull: true, name: 'invoiced_on')  BusinessDate? invoicedOn, @JsonKey(includeIfNull: true, name: 'collected_on')  BusinessDate? collectedOn, @JsonKey(includeIfNull: true, name: 'platform_activity')  PlatformActivity? platformActivity, @JsonKey(includeIfNull: true, name: 'cfdi_uuid')  CfdiUuid? cfdiUuid,  RecordDescription? description, @JsonKey(name: 'pending_collection')  bool pendingCollection,  List<Warning> warnings, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'updated_at')  DateTime updatedAt)  $default,) {final _that = this;
switch (_that) {
case _Income():
return $default(_that.id,_that.taxpayerId,_that.regimeCode,_that.counterpartyKind,_that.operationKind,_that.residentialRent,_that.viaPlatform,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.isrWithheldCents,_that.ivaWithheldCents,_that.withholding,_that.invoicedOn,_that.collectedOn,_that.platformActivity,_that.cfdiUuid,_that.description,_that.pendingCollection,_that.warnings,_that.createdAt,_that.updatedAt);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( Uuid id, @JsonKey(name: 'taxpayer_id')  Uuid taxpayerId, @JsonKey(name: 'regime_code')  RegimeCode regimeCode, @JsonKey(name: 'counterparty_kind')  CounterpartyKind counterpartyKind, @JsonKey(name: 'operation_kind')  OperationKind operationKind, @JsonKey(name: 'residential_rent')  bool residentialRent, @JsonKey(name: 'via_platform')  bool viaPlatform, @JsonKey(name: 'amount_cents')  PositiveCents amountCents, @JsonKey(name: 'iva_rate')  IvaRate ivaRate, @JsonKey(name: 'iva_cents')  NonNegativeCents ivaCents, @JsonKey(name: 'isr_withheld_cents')  NonNegativeCents isrWithheldCents, @JsonKey(name: 'iva_withheld_cents')  NonNegativeCents ivaWithheldCents,  Withholding withholding, @JsonKey(includeIfNull: true, name: 'invoiced_on')  BusinessDate? invoicedOn, @JsonKey(includeIfNull: true, name: 'collected_on')  BusinessDate? collectedOn, @JsonKey(includeIfNull: true, name: 'platform_activity')  PlatformActivity? platformActivity, @JsonKey(includeIfNull: true, name: 'cfdi_uuid')  CfdiUuid? cfdiUuid,  RecordDescription? description, @JsonKey(name: 'pending_collection')  bool pendingCollection,  List<Warning> warnings, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'updated_at')  DateTime updatedAt)?  $default,) {final _that = this;
switch (_that) {
case _Income() when $default != null:
return $default(_that.id,_that.taxpayerId,_that.regimeCode,_that.counterpartyKind,_that.operationKind,_that.residentialRent,_that.viaPlatform,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.isrWithheldCents,_that.ivaWithheldCents,_that.withholding,_that.invoicedOn,_that.collectedOn,_that.platformActivity,_that.cfdiUuid,_that.description,_that.pendingCollection,_that.warnings,_that.createdAt,_that.updatedAt);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _Income implements Income {
  const _Income({required this.id, @JsonKey(name: 'taxpayer_id') required this.taxpayerId, @JsonKey(name: 'regime_code') required this.regimeCode, @JsonKey(name: 'counterparty_kind') required this.counterpartyKind, @JsonKey(name: 'operation_kind') required this.operationKind, @JsonKey(name: 'residential_rent') required this.residentialRent, @JsonKey(name: 'via_platform') required this.viaPlatform, @JsonKey(name: 'amount_cents') required this.amountCents, @JsonKey(name: 'iva_rate') required this.ivaRate, @JsonKey(name: 'iva_cents') required this.ivaCents, @JsonKey(name: 'isr_withheld_cents') required this.isrWithheldCents, @JsonKey(name: 'iva_withheld_cents') required this.ivaWithheldCents, required this.withholding, @JsonKey(includeIfNull: true, name: 'invoiced_on') required this.invoicedOn, @JsonKey(includeIfNull: true, name: 'collected_on') required this.collectedOn, @JsonKey(includeIfNull: true, name: 'platform_activity') required this.platformActivity, @JsonKey(includeIfNull: true, name: 'cfdi_uuid') required this.cfdiUuid, required this.description, @JsonKey(name: 'pending_collection') required this.pendingCollection, required  List<Warning> warnings, @JsonKey(name: 'created_at') required this.createdAt, @JsonKey(name: 'updated_at') required this.updatedAt}): _warnings = warnings;
  factory _Income.fromJson(Map<String, dynamic> json) => _$IncomeFromJson(json);

@override final  Uuid id;
@override@JsonKey(name: 'taxpayer_id') final  Uuid taxpayerId;
@override@JsonKey(name: 'regime_code') final  RegimeCode regimeCode;
/// Who was billed.
@override@JsonKey(name: 'counterparty_kind') final  CounterpartyKind counterpartyKind;
/// As sent, or the default stored when the question was hidden (last used in the regime, else the regime default).
@override@JsonKey(name: 'operation_kind') final  OperationKind operationKind;
@override@JsonKey(name: 'residential_rent') final  bool residentialRent;
@override@JsonKey(name: 'via_platform') final  bool viaPlatform;
@override@JsonKey(name: 'amount_cents') final  PositiveCents amountCents;
@override@JsonKey(name: 'iva_rate') final  IvaRate ivaRate;
@override@JsonKey(name: 'iva_cents') final  NonNegativeCents ivaCents;
/// Read-only mirror of `withholding.isr_cents` (kept for list totals and the estimate).
@override@JsonKey(name: 'isr_withheld_cents') final  NonNegativeCents isrWithheldCents;
/// Read-only mirror of `withholding.iva_cents` (kept for list totals and the estimate).
@override@JsonKey(name: 'iva_withheld_cents') final  NonNegativeCents ivaWithheldCents;
@override final  Withholding withholding;
@override@JsonKey(includeIfNull: true, name: 'invoiced_on') final  BusinessDate? invoicedOn;
@override@JsonKey(includeIfNull: true, name: 'collected_on') final  BusinessDate? collectedOn;
@override@JsonKey(includeIfNull: true, name: 'platform_activity') final  PlatformActivity? platformActivity;
@override@JsonKey(includeIfNull: true, name: 'cfdi_uuid') final  CfdiUuid? cfdiUuid;
@override final  RecordDescription? description;
/// Accrual income with `collected_on` still `null` ("Por cobrar"); counts for IVA only once collected.
@override@JsonKey(name: 'pending_collection') final  bool pendingCollection;
/// Non-blocking notices computed on writes (create, update, restore), never stored; `[]` on reads. Codes: `duplicate_cfdi_uuid` (`cfdi_uuid`), `provisional_parameters` (`collected_on` / `invoiced_on`), `withholding_cpa_unverified` (`withholding`), `foreign_counterparty_not_supported` (`counterparty_kind`). `withholding_differs_from_rate` is retired: never sent since 0.11.3 (no typed amount is left to compare).
 final  List<Warning> _warnings;
/// Non-blocking notices computed on writes (create, update, restore), never stored; `[]` on reads. Codes: `duplicate_cfdi_uuid` (`cfdi_uuid`), `provisional_parameters` (`collected_on` / `invoiced_on`), `withholding_cpa_unverified` (`withholding`), `foreign_counterparty_not_supported` (`counterparty_kind`). `withholding_differs_from_rate` is retired: never sent since 0.11.3 (no typed amount is left to compare).
@override List<Warning> get warnings {
  if (_warnings is EqualUnmodifiableListView) return _warnings;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_warnings);
}

@override@JsonKey(name: 'created_at') final  DateTime createdAt;
@override@JsonKey(name: 'updated_at') final  DateTime updatedAt;

/// Create a copy of Income
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$IncomeCopyWith<_Income> get copyWith => __$IncomeCopyWithImpl<_Income>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$IncomeToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _Income&&(identical(other.id, id) || other.id == id)&&(identical(other.taxpayerId, taxpayerId) || other.taxpayerId == taxpayerId)&&(identical(other.regimeCode, regimeCode) || other.regimeCode == regimeCode)&&(identical(other.counterpartyKind, counterpartyKind) || other.counterpartyKind == counterpartyKind)&&(identical(other.operationKind, operationKind) || other.operationKind == operationKind)&&(identical(other.residentialRent, residentialRent) || other.residentialRent == residentialRent)&&(identical(other.viaPlatform, viaPlatform) || other.viaPlatform == viaPlatform)&&(identical(other.amountCents, amountCents) || other.amountCents == amountCents)&&(identical(other.ivaRate, ivaRate) || other.ivaRate == ivaRate)&&(identical(other.ivaCents, ivaCents) || other.ivaCents == ivaCents)&&(identical(other.isrWithheldCents, isrWithheldCents) || other.isrWithheldCents == isrWithheldCents)&&(identical(other.ivaWithheldCents, ivaWithheldCents) || other.ivaWithheldCents == ivaWithheldCents)&&(identical(other.withholding, withholding) || other.withholding == withholding)&&(identical(other.invoicedOn, invoicedOn) || other.invoicedOn == invoicedOn)&&(identical(other.collectedOn, collectedOn) || other.collectedOn == collectedOn)&&(identical(other.platformActivity, platformActivity) || other.platformActivity == platformActivity)&&(identical(other.cfdiUuid, cfdiUuid) || other.cfdiUuid == cfdiUuid)&&(identical(other.description, description) || other.description == description)&&(identical(other.pendingCollection, pendingCollection) || other.pendingCollection == pendingCollection)&&const DeepCollectionEquality().equals(other.warnings, _warnings)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.updatedAt, updatedAt) || other.updatedAt == updatedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hashAll([runtimeType,id,taxpayerId,regimeCode,counterpartyKind,operationKind,residentialRent,viaPlatform,amountCents,ivaRate,ivaCents,isrWithheldCents,ivaWithheldCents,withholding,invoicedOn,collectedOn,platformActivity,cfdiUuid,description,pendingCollection,const DeepCollectionEquality().hash(_warnings),createdAt,updatedAt]);
}

@override
String toString() {
    return 'Income(id: $id, taxpayerId: $taxpayerId, regimeCode: $regimeCode, counterpartyKind: $counterpartyKind, operationKind: $operationKind, residentialRent: $residentialRent, viaPlatform: $viaPlatform, amountCents: $amountCents, ivaRate: $ivaRate, ivaCents: $ivaCents, isrWithheldCents: $isrWithheldCents, ivaWithheldCents: $ivaWithheldCents, withholding: $withholding, invoicedOn: $invoicedOn, collectedOn: $collectedOn, platformActivity: $platformActivity, cfdiUuid: $cfdiUuid, description: $description, pendingCollection: $pendingCollection, warnings: $warnings, createdAt: $createdAt, updatedAt: $updatedAt)';
}


}

/// @nodoc
abstract mixin class _$IncomeCopyWith<$Res> implements $IncomeCopyWith<$Res> {
  factory _$IncomeCopyWith(_Income value, $Res Function(_Income) _then) = __$IncomeCopyWithImpl;
@override @useResult
$Res call({
 Uuid id,@JsonKey(name: 'taxpayer_id') Uuid taxpayerId,@JsonKey(name: 'regime_code') RegimeCode regimeCode,@JsonKey(name: 'counterparty_kind') CounterpartyKind counterpartyKind,@JsonKey(name: 'operation_kind') OperationKind operationKind,@JsonKey(name: 'residential_rent') bool residentialRent,@JsonKey(name: 'via_platform') bool viaPlatform,@JsonKey(name: 'amount_cents') PositiveCents amountCents,@JsonKey(name: 'iva_rate') IvaRate ivaRate,@JsonKey(name: 'iva_cents') NonNegativeCents ivaCents,@JsonKey(name: 'isr_withheld_cents') NonNegativeCents isrWithheldCents,@JsonKey(name: 'iva_withheld_cents') NonNegativeCents ivaWithheldCents, Withholding withholding,@JsonKey(includeIfNull: true, name: 'invoiced_on') BusinessDate? invoicedOn,@JsonKey(includeIfNull: true, name: 'collected_on') BusinessDate? collectedOn,@JsonKey(includeIfNull: true, name: 'platform_activity') PlatformActivity? platformActivity,@JsonKey(includeIfNull: true, name: 'cfdi_uuid') CfdiUuid? cfdiUuid, RecordDescription? description,@JsonKey(name: 'pending_collection') bool pendingCollection, List<Warning> warnings,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'updated_at') DateTime updatedAt
});


@override $WithholdingCopyWith<$Res> get withholding;

}
/// @nodoc
class __$IncomeCopyWithImpl<$Res>
    implements _$IncomeCopyWith<$Res> {
  __$IncomeCopyWithImpl(this._self, this._then);

  final _Income _self;
  final $Res Function(_Income) _then;

/// Create a copy of Income
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? taxpayerId = null,Object? regimeCode = null,Object? counterpartyKind = null,Object? operationKind = null,Object? residentialRent = null,Object? viaPlatform = null,Object? amountCents = null,Object? ivaRate = null,Object? ivaCents = null,Object? isrWithheldCents = null,Object? ivaWithheldCents = null,Object? withholding = null,Object? invoicedOn = freezed,Object? collectedOn = freezed,Object? platformActivity = freezed,Object? cfdiUuid = freezed,Object? description = freezed,Object? pendingCollection = null,Object? warnings = null,Object? createdAt = null,Object? updatedAt = null,}) {
  return _then(_Income(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid,taxpayerId: null == taxpayerId ? _self.taxpayerId : taxpayerId // ignore: cast_nullable_to_non_nullable
as Uuid,regimeCode: null == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode,counterpartyKind: null == counterpartyKind ? _self.counterpartyKind : counterpartyKind // ignore: cast_nullable_to_non_nullable
as CounterpartyKind,operationKind: null == operationKind ? _self.operationKind : operationKind // ignore: cast_nullable_to_non_nullable
as OperationKind,residentialRent: null == residentialRent ? _self.residentialRent : residentialRent // ignore: cast_nullable_to_non_nullable
as bool,viaPlatform: null == viaPlatform ? _self.viaPlatform : viaPlatform // ignore: cast_nullable_to_non_nullable
as bool,amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as PositiveCents,ivaRate: null == ivaRate ? _self.ivaRate : ivaRate // ignore: cast_nullable_to_non_nullable
as IvaRate,ivaCents: null == ivaCents ? _self.ivaCents : ivaCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,isrWithheldCents: null == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,ivaWithheldCents: null == ivaWithheldCents ? _self.ivaWithheldCents : ivaWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,withholding: null == withholding ? _self.withholding : withholding // ignore: cast_nullable_to_non_nullable
as Withholding,invoicedOn: freezed == invoicedOn ? _self.invoicedOn : invoicedOn // ignore: cast_nullable_to_non_nullable
as BusinessDate?,collectedOn: freezed == collectedOn ? _self.collectedOn : collectedOn // ignore: cast_nullable_to_non_nullable
as BusinessDate?,platformActivity: freezed == platformActivity ? _self.platformActivity : platformActivity // ignore: cast_nullable_to_non_nullable
as PlatformActivity?,cfdiUuid: freezed == cfdiUuid ? _self.cfdiUuid : cfdiUuid // ignore: cast_nullable_to_non_nullable
as CfdiUuid?,description: freezed == description ? _self.description : description // ignore: cast_nullable_to_non_nullable
as RecordDescription?,pendingCollection: null == pendingCollection ? _self.pendingCollection : pendingCollection // ignore: cast_nullable_to_non_nullable
as bool,warnings: null == warnings ? _self._warnings : warnings // ignore: cast_nullable_to_non_nullable
as List<Warning>,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,updatedAt: null == updatedAt ? _self.updatedAt : updatedAt // ignore: cast_nullable_to_non_nullable
as DateTime,
  ));
}

/// Create a copy of Income
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$WithholdingCopyWith<$Res> get withholding {
  
  return $WithholdingCopyWith<$Res>(_self.withholding, (value) {
    return _then(_self.copyWith(withholding: value));
  });
}
}

// dart format on
