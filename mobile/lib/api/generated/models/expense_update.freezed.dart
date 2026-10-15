// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'expense_update.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$ExpenseUpdate {

@JsonKey(includeIfNull: false, name: 'regime_code') RegimeCode? get regimeCode;@JsonKey(includeIfNull: false, name: 'category_code') String? get categoryCode;@JsonKey(includeIfNull: false, name: 'amount_cents') PositiveCents? get amountCents;@JsonKey(includeIfNull: false, name: 'iva_rate') IvaRate? get ivaRate;@JsonKey(includeIfNull: false, name: 'iva_cents') NonNegativeCents? get ivaCents;@JsonKey(includeIfNull: false, name: 'payment_method') PaymentMethod? get paymentMethod;@JsonKey(includeIfNull: false, name: 'paid_on') BusinessDate? get paidOn;@JsonKey(includeIfNull: false, name: 'has_cfdi') bool? get hasCfdi;@JsonKey(includeIfNull: false, name: 'cfdi_uuid') CfdiUuid? get cfdiUuid;@JsonKey(includeIfNull: false) bool? get deductible;@JsonKey(includeIfNull: false, name: 'supplier_kind') SupplierKind? get supplierKind;@JsonKey(includeIfNull: false, name: 'supplier_regime') SupplierRegime? get supplierRegime;@JsonKey(includeIfNull: false, name: 'operation_kind') OperationKind? get operationKind;@JsonKey(includeIfNull: false, name: 'residential_rent') bool? get residentialRent;/// Retired by WH-07 (0.11.3). Any value answers `400 additionalProperties`.
@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') NonNegativeCents? get isrWithheldCents;/// Retired by WH-07 (0.11.3). Any value answers `400 additionalProperties`.
@JsonKey(includeIfNull: false, name: 'iva_withheld_cents') NonNegativeCents? get ivaWithheldCents;@JsonKey(includeIfNull: false) RecordDescription? get description;
/// Create a copy of ExpenseUpdate
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ExpenseUpdateCopyWith<ExpenseUpdate> get copyWith => _$ExpenseUpdateCopyWithImpl<ExpenseUpdate>(this as ExpenseUpdate, _$identity);

  /// Serializes this ExpenseUpdate to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as ExpenseUpdate;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is ExpenseUpdate&&(identical(other.regimeCode, _this.regimeCode) || other.regimeCode == _this.regimeCode)&&(identical(other.categoryCode, _this.categoryCode) || other.categoryCode == _this.categoryCode)&&(identical(other.amountCents, _this.amountCents) || other.amountCents == _this.amountCents)&&(identical(other.ivaRate, _this.ivaRate) || other.ivaRate == _this.ivaRate)&&(identical(other.ivaCents, _this.ivaCents) || other.ivaCents == _this.ivaCents)&&(identical(other.paymentMethod, _this.paymentMethod) || other.paymentMethod == _this.paymentMethod)&&(identical(other.paidOn, _this.paidOn) || other.paidOn == _this.paidOn)&&(identical(other.hasCfdi, _this.hasCfdi) || other.hasCfdi == _this.hasCfdi)&&(identical(other.cfdiUuid, _this.cfdiUuid) || other.cfdiUuid == _this.cfdiUuid)&&(identical(other.deductible, _this.deductible) || other.deductible == _this.deductible)&&(identical(other.supplierKind, _this.supplierKind) || other.supplierKind == _this.supplierKind)&&(identical(other.supplierRegime, _this.supplierRegime) || other.supplierRegime == _this.supplierRegime)&&(identical(other.operationKind, _this.operationKind) || other.operationKind == _this.operationKind)&&(identical(other.residentialRent, _this.residentialRent) || other.residentialRent == _this.residentialRent)&&(identical(other.isrWithheldCents, _this.isrWithheldCents) || other.isrWithheldCents == _this.isrWithheldCents)&&(identical(other.ivaWithheldCents, _this.ivaWithheldCents) || other.ivaWithheldCents == _this.ivaWithheldCents)&&(identical(other.description, _this.description) || other.description == _this.description));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as ExpenseUpdate;
  return Object.hash(runtimeType,_this.regimeCode,_this.categoryCode,_this.amountCents,_this.ivaRate,_this.ivaCents,_this.paymentMethod,_this.paidOn,_this.hasCfdi,_this.cfdiUuid,_this.deductible,_this.supplierKind,_this.supplierRegime,_this.operationKind,_this.residentialRent,_this.isrWithheldCents,_this.ivaWithheldCents,_this.description);
}

@override
String toString() {
  final _this = this as ExpenseUpdate;
  return 'ExpenseUpdate(regimeCode: ${_this.regimeCode}, categoryCode: ${_this.categoryCode}, amountCents: ${_this.amountCents}, ivaRate: ${_this.ivaRate}, ivaCents: ${_this.ivaCents}, paymentMethod: ${_this.paymentMethod}, paidOn: ${_this.paidOn}, hasCfdi: ${_this.hasCfdi}, cfdiUuid: ${_this.cfdiUuid}, deductible: ${_this.deductible}, supplierKind: ${_this.supplierKind}, supplierRegime: ${_this.supplierRegime}, operationKind: ${_this.operationKind}, residentialRent: ${_this.residentialRent}, isrWithheldCents: ${_this.isrWithheldCents}, ivaWithheldCents: ${_this.ivaWithheldCents}, description: ${_this.description})';
}


}

/// @nodoc
abstract mixin class $ExpenseUpdateCopyWith<$Res>  {
  factory $ExpenseUpdateCopyWith(ExpenseUpdate value, $Res Function(ExpenseUpdate) _then) = _$ExpenseUpdateCopyWithImpl;
@useResult
$Res call({
@JsonKey(includeIfNull: false, name: 'regime_code') RegimeCode? regimeCode,@JsonKey(includeIfNull: false, name: 'category_code') String? categoryCode,@JsonKey(includeIfNull: false, name: 'amount_cents') PositiveCents? amountCents,@JsonKey(includeIfNull: false, name: 'iva_rate') IvaRate? ivaRate,@JsonKey(includeIfNull: false, name: 'iva_cents') NonNegativeCents? ivaCents,@JsonKey(includeIfNull: false, name: 'payment_method') PaymentMethod? paymentMethod,@JsonKey(includeIfNull: false, name: 'paid_on') BusinessDate? paidOn,@JsonKey(includeIfNull: false, name: 'has_cfdi') bool? hasCfdi,@JsonKey(includeIfNull: false, name: 'cfdi_uuid') CfdiUuid? cfdiUuid,@JsonKey(includeIfNull: false) bool? deductible,@JsonKey(includeIfNull: false, name: 'supplier_kind') SupplierKind? supplierKind,@JsonKey(includeIfNull: false, name: 'supplier_regime') SupplierRegime? supplierRegime,@JsonKey(includeIfNull: false, name: 'operation_kind') OperationKind? operationKind,@JsonKey(includeIfNull: false, name: 'residential_rent') bool? residentialRent,@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') NonNegativeCents? isrWithheldCents,@JsonKey(includeIfNull: false, name: 'iva_withheld_cents') NonNegativeCents? ivaWithheldCents,@JsonKey(includeIfNull: false) RecordDescription? description
});




}
/// @nodoc
class _$ExpenseUpdateCopyWithImpl<$Res>
    implements $ExpenseUpdateCopyWith<$Res> {
  _$ExpenseUpdateCopyWithImpl(this._self, this._then);

  final ExpenseUpdate _self;
  final $Res Function(ExpenseUpdate) _then;

/// Create a copy of ExpenseUpdate
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? regimeCode = freezed,Object? categoryCode = freezed,Object? amountCents = freezed,Object? ivaRate = freezed,Object? ivaCents = freezed,Object? paymentMethod = freezed,Object? paidOn = freezed,Object? hasCfdi = freezed,Object? cfdiUuid = freezed,Object? deductible = freezed,Object? supplierKind = freezed,Object? supplierRegime = freezed,Object? operationKind = freezed,Object? residentialRent = freezed,Object? isrWithheldCents = freezed,Object? ivaWithheldCents = freezed,Object? description = freezed,}) {
  return _then(ExpenseUpdate(
regimeCode: freezed == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode?,categoryCode: freezed == categoryCode ? _self.categoryCode : categoryCode // ignore: cast_nullable_to_non_nullable
as String?,amountCents: freezed == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as PositiveCents?,ivaRate: freezed == ivaRate ? _self.ivaRate : ivaRate // ignore: cast_nullable_to_non_nullable
as IvaRate?,ivaCents: freezed == ivaCents ? _self.ivaCents : ivaCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,paymentMethod: freezed == paymentMethod ? _self.paymentMethod : paymentMethod // ignore: cast_nullable_to_non_nullable
as PaymentMethod?,paidOn: freezed == paidOn ? _self.paidOn : paidOn // ignore: cast_nullable_to_non_nullable
as BusinessDate?,hasCfdi: freezed == hasCfdi ? _self.hasCfdi : hasCfdi // ignore: cast_nullable_to_non_nullable
as bool?,cfdiUuid: freezed == cfdiUuid ? _self.cfdiUuid : cfdiUuid // ignore: cast_nullable_to_non_nullable
as CfdiUuid?,deductible: freezed == deductible ? _self.deductible : deductible // ignore: cast_nullable_to_non_nullable
as bool?,supplierKind: freezed == supplierKind ? _self.supplierKind : supplierKind // ignore: cast_nullable_to_non_nullable
as SupplierKind?,supplierRegime: freezed == supplierRegime ? _self.supplierRegime : supplierRegime // ignore: cast_nullable_to_non_nullable
as SupplierRegime?,operationKind: freezed == operationKind ? _self.operationKind : operationKind // ignore: cast_nullable_to_non_nullable
as OperationKind?,residentialRent: freezed == residentialRent ? _self.residentialRent : residentialRent // ignore: cast_nullable_to_non_nullable
as bool?,isrWithheldCents: freezed == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,ivaWithheldCents: freezed == ivaWithheldCents ? _self.ivaWithheldCents : ivaWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,description: freezed == description ? _self.description : description // ignore: cast_nullable_to_non_nullable
as RecordDescription?,
  ));
}

}


/// Adds pattern-matching-related methods to [ExpenseUpdate].
extension ExpenseUpdatePatterns on ExpenseUpdate {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _ExpenseUpdate value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _ExpenseUpdate() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _ExpenseUpdate value)  $default,){
final _that = this;
switch (_that) {
case _ExpenseUpdate():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _ExpenseUpdate value)?  $default,){
final _that = this;
switch (_that) {
case _ExpenseUpdate() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(includeIfNull: false, name: 'regime_code')  RegimeCode? regimeCode, @JsonKey(includeIfNull: false, name: 'category_code')  String? categoryCode, @JsonKey(includeIfNull: false, name: 'amount_cents')  PositiveCents? amountCents, @JsonKey(includeIfNull: false, name: 'iva_rate')  IvaRate? ivaRate, @JsonKey(includeIfNull: false, name: 'iva_cents')  NonNegativeCents? ivaCents, @JsonKey(includeIfNull: false, name: 'payment_method')  PaymentMethod? paymentMethod, @JsonKey(includeIfNull: false, name: 'paid_on')  BusinessDate? paidOn, @JsonKey(includeIfNull: false, name: 'has_cfdi')  bool? hasCfdi, @JsonKey(includeIfNull: false, name: 'cfdi_uuid')  CfdiUuid? cfdiUuid, @JsonKey(includeIfNull: false)  bool? deductible, @JsonKey(includeIfNull: false, name: 'supplier_kind')  SupplierKind? supplierKind, @JsonKey(includeIfNull: false, name: 'supplier_regime')  SupplierRegime? supplierRegime, @JsonKey(includeIfNull: false, name: 'operation_kind')  OperationKind? operationKind, @JsonKey(includeIfNull: false, name: 'residential_rent')  bool? residentialRent, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(includeIfNull: false, name: 'iva_withheld_cents')  NonNegativeCents? ivaWithheldCents, @JsonKey(includeIfNull: false)  RecordDescription? description)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _ExpenseUpdate() when $default != null:
return $default(_that.regimeCode,_that.categoryCode,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.paymentMethod,_that.paidOn,_that.hasCfdi,_that.cfdiUuid,_that.deductible,_that.supplierKind,_that.supplierRegime,_that.operationKind,_that.residentialRent,_that.isrWithheldCents,_that.ivaWithheldCents,_that.description);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(includeIfNull: false, name: 'regime_code')  RegimeCode? regimeCode, @JsonKey(includeIfNull: false, name: 'category_code')  String? categoryCode, @JsonKey(includeIfNull: false, name: 'amount_cents')  PositiveCents? amountCents, @JsonKey(includeIfNull: false, name: 'iva_rate')  IvaRate? ivaRate, @JsonKey(includeIfNull: false, name: 'iva_cents')  NonNegativeCents? ivaCents, @JsonKey(includeIfNull: false, name: 'payment_method')  PaymentMethod? paymentMethod, @JsonKey(includeIfNull: false, name: 'paid_on')  BusinessDate? paidOn, @JsonKey(includeIfNull: false, name: 'has_cfdi')  bool? hasCfdi, @JsonKey(includeIfNull: false, name: 'cfdi_uuid')  CfdiUuid? cfdiUuid, @JsonKey(includeIfNull: false)  bool? deductible, @JsonKey(includeIfNull: false, name: 'supplier_kind')  SupplierKind? supplierKind, @JsonKey(includeIfNull: false, name: 'supplier_regime')  SupplierRegime? supplierRegime, @JsonKey(includeIfNull: false, name: 'operation_kind')  OperationKind? operationKind, @JsonKey(includeIfNull: false, name: 'residential_rent')  bool? residentialRent, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(includeIfNull: false, name: 'iva_withheld_cents')  NonNegativeCents? ivaWithheldCents, @JsonKey(includeIfNull: false)  RecordDescription? description)  $default,) {final _that = this;
switch (_that) {
case _ExpenseUpdate():
return $default(_that.regimeCode,_that.categoryCode,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.paymentMethod,_that.paidOn,_that.hasCfdi,_that.cfdiUuid,_that.deductible,_that.supplierKind,_that.supplierRegime,_that.operationKind,_that.residentialRent,_that.isrWithheldCents,_that.ivaWithheldCents,_that.description);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(includeIfNull: false, name: 'regime_code')  RegimeCode? regimeCode, @JsonKey(includeIfNull: false, name: 'category_code')  String? categoryCode, @JsonKey(includeIfNull: false, name: 'amount_cents')  PositiveCents? amountCents, @JsonKey(includeIfNull: false, name: 'iva_rate')  IvaRate? ivaRate, @JsonKey(includeIfNull: false, name: 'iva_cents')  NonNegativeCents? ivaCents, @JsonKey(includeIfNull: false, name: 'payment_method')  PaymentMethod? paymentMethod, @JsonKey(includeIfNull: false, name: 'paid_on')  BusinessDate? paidOn, @JsonKey(includeIfNull: false, name: 'has_cfdi')  bool? hasCfdi, @JsonKey(includeIfNull: false, name: 'cfdi_uuid')  CfdiUuid? cfdiUuid, @JsonKey(includeIfNull: false)  bool? deductible, @JsonKey(includeIfNull: false, name: 'supplier_kind')  SupplierKind? supplierKind, @JsonKey(includeIfNull: false, name: 'supplier_regime')  SupplierRegime? supplierRegime, @JsonKey(includeIfNull: false, name: 'operation_kind')  OperationKind? operationKind, @JsonKey(includeIfNull: false, name: 'residential_rent')  bool? residentialRent, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(includeIfNull: false, name: 'iva_withheld_cents')  NonNegativeCents? ivaWithheldCents, @JsonKey(includeIfNull: false)  RecordDescription? description)?  $default,) {final _that = this;
switch (_that) {
case _ExpenseUpdate() when $default != null:
return $default(_that.regimeCode,_that.categoryCode,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.paymentMethod,_that.paidOn,_that.hasCfdi,_that.cfdiUuid,_that.deductible,_that.supplierKind,_that.supplierRegime,_that.operationKind,_that.residentialRent,_that.isrWithheldCents,_that.ivaWithheldCents,_that.description);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _ExpenseUpdate implements ExpenseUpdate {
  const _ExpenseUpdate({@JsonKey(includeIfNull: false, name: 'regime_code') this.regimeCode, @JsonKey(includeIfNull: false, name: 'category_code') this.categoryCode, @JsonKey(includeIfNull: false, name: 'amount_cents') this.amountCents, @JsonKey(includeIfNull: false, name: 'iva_rate') this.ivaRate, @JsonKey(includeIfNull: false, name: 'iva_cents') this.ivaCents, @JsonKey(includeIfNull: false, name: 'payment_method') this.paymentMethod, @JsonKey(includeIfNull: false, name: 'paid_on') this.paidOn, @JsonKey(includeIfNull: false, name: 'has_cfdi') this.hasCfdi, @JsonKey(includeIfNull: false, name: 'cfdi_uuid') this.cfdiUuid, @JsonKey(includeIfNull: false) this.deductible, @JsonKey(includeIfNull: false, name: 'supplier_kind') this.supplierKind, @JsonKey(includeIfNull: false, name: 'supplier_regime') this.supplierRegime, @JsonKey(includeIfNull: false, name: 'operation_kind') this.operationKind, @JsonKey(includeIfNull: false, name: 'residential_rent') this.residentialRent, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents') this.isrWithheldCents, @JsonKey(includeIfNull: false, name: 'iva_withheld_cents') this.ivaWithheldCents, @JsonKey(includeIfNull: false) this.description});
  factory _ExpenseUpdate.fromJson(Map<String, dynamic> json) => _$ExpenseUpdateFromJson(json);

@override@JsonKey(includeIfNull: false, name: 'regime_code') final  RegimeCode? regimeCode;
@override@JsonKey(includeIfNull: false, name: 'category_code') final  String? categoryCode;
@override@JsonKey(includeIfNull: false, name: 'amount_cents') final  PositiveCents? amountCents;
@override@JsonKey(includeIfNull: false, name: 'iva_rate') final  IvaRate? ivaRate;
@override@JsonKey(includeIfNull: false, name: 'iva_cents') final  NonNegativeCents? ivaCents;
@override@JsonKey(includeIfNull: false, name: 'payment_method') final  PaymentMethod? paymentMethod;
@override@JsonKey(includeIfNull: false, name: 'paid_on') final  BusinessDate? paidOn;
@override@JsonKey(includeIfNull: false, name: 'has_cfdi') final  bool? hasCfdi;
@override@JsonKey(includeIfNull: false, name: 'cfdi_uuid') final  CfdiUuid? cfdiUuid;
@override@JsonKey(includeIfNull: false) final  bool? deductible;
@override@JsonKey(includeIfNull: false, name: 'supplier_kind') final  SupplierKind? supplierKind;
@override@JsonKey(includeIfNull: false, name: 'supplier_regime') final  SupplierRegime? supplierRegime;
@override@JsonKey(includeIfNull: false, name: 'operation_kind') final  OperationKind? operationKind;
@override@JsonKey(includeIfNull: false, name: 'residential_rent') final  bool? residentialRent;
/// Retired by WH-07 (0.11.3). Any value answers `400 additionalProperties`.
@override@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') final  NonNegativeCents? isrWithheldCents;
/// Retired by WH-07 (0.11.3). Any value answers `400 additionalProperties`.
@override@JsonKey(includeIfNull: false, name: 'iva_withheld_cents') final  NonNegativeCents? ivaWithheldCents;
@override@JsonKey(includeIfNull: false) final  RecordDescription? description;

/// Create a copy of ExpenseUpdate
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ExpenseUpdateCopyWith<_ExpenseUpdate> get copyWith => __$ExpenseUpdateCopyWithImpl<_ExpenseUpdate>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$ExpenseUpdateToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _ExpenseUpdate&&(identical(other.regimeCode, regimeCode) || other.regimeCode == regimeCode)&&(identical(other.categoryCode, categoryCode) || other.categoryCode == categoryCode)&&(identical(other.amountCents, amountCents) || other.amountCents == amountCents)&&(identical(other.ivaRate, ivaRate) || other.ivaRate == ivaRate)&&(identical(other.ivaCents, ivaCents) || other.ivaCents == ivaCents)&&(identical(other.paymentMethod, paymentMethod) || other.paymentMethod == paymentMethod)&&(identical(other.paidOn, paidOn) || other.paidOn == paidOn)&&(identical(other.hasCfdi, hasCfdi) || other.hasCfdi == hasCfdi)&&(identical(other.cfdiUuid, cfdiUuid) || other.cfdiUuid == cfdiUuid)&&(identical(other.deductible, deductible) || other.deductible == deductible)&&(identical(other.supplierKind, supplierKind) || other.supplierKind == supplierKind)&&(identical(other.supplierRegime, supplierRegime) || other.supplierRegime == supplierRegime)&&(identical(other.operationKind, operationKind) || other.operationKind == operationKind)&&(identical(other.residentialRent, residentialRent) || other.residentialRent == residentialRent)&&(identical(other.isrWithheldCents, isrWithheldCents) || other.isrWithheldCents == isrWithheldCents)&&(identical(other.ivaWithheldCents, ivaWithheldCents) || other.ivaWithheldCents == ivaWithheldCents)&&(identical(other.description, description) || other.description == description));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,regimeCode,categoryCode,amountCents,ivaRate,ivaCents,paymentMethod,paidOn,hasCfdi,cfdiUuid,deductible,supplierKind,supplierRegime,operationKind,residentialRent,isrWithheldCents,ivaWithheldCents,description);
}

@override
String toString() {
    return 'ExpenseUpdate(regimeCode: $regimeCode, categoryCode: $categoryCode, amountCents: $amountCents, ivaRate: $ivaRate, ivaCents: $ivaCents, paymentMethod: $paymentMethod, paidOn: $paidOn, hasCfdi: $hasCfdi, cfdiUuid: $cfdiUuid, deductible: $deductible, supplierKind: $supplierKind, supplierRegime: $supplierRegime, operationKind: $operationKind, residentialRent: $residentialRent, isrWithheldCents: $isrWithheldCents, ivaWithheldCents: $ivaWithheldCents, description: $description)';
}


}

/// @nodoc
abstract mixin class _$ExpenseUpdateCopyWith<$Res> implements $ExpenseUpdateCopyWith<$Res> {
  factory _$ExpenseUpdateCopyWith(_ExpenseUpdate value, $Res Function(_ExpenseUpdate) _then) = __$ExpenseUpdateCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(includeIfNull: false, name: 'regime_code') RegimeCode? regimeCode,@JsonKey(includeIfNull: false, name: 'category_code') String? categoryCode,@JsonKey(includeIfNull: false, name: 'amount_cents') PositiveCents? amountCents,@JsonKey(includeIfNull: false, name: 'iva_rate') IvaRate? ivaRate,@JsonKey(includeIfNull: false, name: 'iva_cents') NonNegativeCents? ivaCents,@JsonKey(includeIfNull: false, name: 'payment_method') PaymentMethod? paymentMethod,@JsonKey(includeIfNull: false, name: 'paid_on') BusinessDate? paidOn,@JsonKey(includeIfNull: false, name: 'has_cfdi') bool? hasCfdi,@JsonKey(includeIfNull: false, name: 'cfdi_uuid') CfdiUuid? cfdiUuid,@JsonKey(includeIfNull: false) bool? deductible,@JsonKey(includeIfNull: false, name: 'supplier_kind') SupplierKind? supplierKind,@JsonKey(includeIfNull: false, name: 'supplier_regime') SupplierRegime? supplierRegime,@JsonKey(includeIfNull: false, name: 'operation_kind') OperationKind? operationKind,@JsonKey(includeIfNull: false, name: 'residential_rent') bool? residentialRent,@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') NonNegativeCents? isrWithheldCents,@JsonKey(includeIfNull: false, name: 'iva_withheld_cents') NonNegativeCents? ivaWithheldCents,@JsonKey(includeIfNull: false) RecordDescription? description
});




}
/// @nodoc
class __$ExpenseUpdateCopyWithImpl<$Res>
    implements _$ExpenseUpdateCopyWith<$Res> {
  __$ExpenseUpdateCopyWithImpl(this._self, this._then);

  final _ExpenseUpdate _self;
  final $Res Function(_ExpenseUpdate) _then;

/// Create a copy of ExpenseUpdate
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? regimeCode = freezed,Object? categoryCode = freezed,Object? amountCents = freezed,Object? ivaRate = freezed,Object? ivaCents = freezed,Object? paymentMethod = freezed,Object? paidOn = freezed,Object? hasCfdi = freezed,Object? cfdiUuid = freezed,Object? deductible = freezed,Object? supplierKind = freezed,Object? supplierRegime = freezed,Object? operationKind = freezed,Object? residentialRent = freezed,Object? isrWithheldCents = freezed,Object? ivaWithheldCents = freezed,Object? description = freezed,}) {
  return _then(_ExpenseUpdate(
regimeCode: freezed == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode?,categoryCode: freezed == categoryCode ? _self.categoryCode : categoryCode // ignore: cast_nullable_to_non_nullable
as String?,amountCents: freezed == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as PositiveCents?,ivaRate: freezed == ivaRate ? _self.ivaRate : ivaRate // ignore: cast_nullable_to_non_nullable
as IvaRate?,ivaCents: freezed == ivaCents ? _self.ivaCents : ivaCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,paymentMethod: freezed == paymentMethod ? _self.paymentMethod : paymentMethod // ignore: cast_nullable_to_non_nullable
as PaymentMethod?,paidOn: freezed == paidOn ? _self.paidOn : paidOn // ignore: cast_nullable_to_non_nullable
as BusinessDate?,hasCfdi: freezed == hasCfdi ? _self.hasCfdi : hasCfdi // ignore: cast_nullable_to_non_nullable
as bool?,cfdiUuid: freezed == cfdiUuid ? _self.cfdiUuid : cfdiUuid // ignore: cast_nullable_to_non_nullable
as CfdiUuid?,deductible: freezed == deductible ? _self.deductible : deductible // ignore: cast_nullable_to_non_nullable
as bool?,supplierKind: freezed == supplierKind ? _self.supplierKind : supplierKind // ignore: cast_nullable_to_non_nullable
as SupplierKind?,supplierRegime: freezed == supplierRegime ? _self.supplierRegime : supplierRegime // ignore: cast_nullable_to_non_nullable
as SupplierRegime?,operationKind: freezed == operationKind ? _self.operationKind : operationKind // ignore: cast_nullable_to_non_nullable
as OperationKind?,residentialRent: freezed == residentialRent ? _self.residentialRent : residentialRent // ignore: cast_nullable_to_non_nullable
as bool?,isrWithheldCents: freezed == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,ivaWithheldCents: freezed == ivaWithheldCents ? _self.ivaWithheldCents : ivaWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,description: freezed == description ? _self.description : description // ignore: cast_nullable_to_non_nullable
as RecordDescription?,
  ));
}


}

// dart format on
