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

@JsonKey(name: 'regime_code') RegimeCode? get regimeCode; String? get description;@JsonKey(name: 'counterparty_type') CounterpartyType? get counterpartyType;/// Subtotal before IVA.
@JsonKey(name: 'amount_cents') NonNegativeCents? get amountCents;@JsonKey(name: 'iva_rate') IvaRate? get ivaRate;@JsonKey(name: 'iva_cents') NonNegativeCents? get ivaCents;@JsonKey(name: 'isr_withheld_cents') NonNegativeCents? get isrWithheldCents;@JsonKey(name: 'iva_withheld_cents') NonNegativeCents? get ivaWithheldCents;@JsonKey(name: 'invoiced_on') DateTime? get invoicedOn;/// Required for cash-basis regimes; may be null for PM 601 (pending collection).
@JsonKey(name: 'collected_on') DateTime? get collectedOn; Uuid? get id;
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
  return identical(this, other) || (other.runtimeType == runtimeType&&other is IncomeCreate&&(identical(other.regimeCode, _this.regimeCode) || other.regimeCode == _this.regimeCode)&&(identical(other.description, _this.description) || other.description == _this.description)&&(identical(other.counterpartyType, _this.counterpartyType) || other.counterpartyType == _this.counterpartyType)&&(identical(other.amountCents, _this.amountCents) || other.amountCents == _this.amountCents)&&(identical(other.ivaRate, _this.ivaRate) || other.ivaRate == _this.ivaRate)&&(identical(other.ivaCents, _this.ivaCents) || other.ivaCents == _this.ivaCents)&&(identical(other.isrWithheldCents, _this.isrWithheldCents) || other.isrWithheldCents == _this.isrWithheldCents)&&(identical(other.ivaWithheldCents, _this.ivaWithheldCents) || other.ivaWithheldCents == _this.ivaWithheldCents)&&(identical(other.invoicedOn, _this.invoicedOn) || other.invoicedOn == _this.invoicedOn)&&(identical(other.collectedOn, _this.collectedOn) || other.collectedOn == _this.collectedOn)&&(identical(other.id, _this.id) || other.id == _this.id));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as IncomeCreate;
  return Object.hash(runtimeType,_this.regimeCode,_this.description,_this.counterpartyType,_this.amountCents,_this.ivaRate,_this.ivaCents,_this.isrWithheldCents,_this.ivaWithheldCents,_this.invoicedOn,_this.collectedOn,_this.id);
}

@override
String toString() {
  final _this = this as IncomeCreate;
  return 'IncomeCreate(regimeCode: ${_this.regimeCode}, description: ${_this.description}, counterpartyType: ${_this.counterpartyType}, amountCents: ${_this.amountCents}, ivaRate: ${_this.ivaRate}, ivaCents: ${_this.ivaCents}, isrWithheldCents: ${_this.isrWithheldCents}, ivaWithheldCents: ${_this.ivaWithheldCents}, invoicedOn: ${_this.invoicedOn}, collectedOn: ${_this.collectedOn}, id: ${_this.id})';
}


}

/// @nodoc
abstract mixin class $IncomeCreateCopyWith<$Res>  {
  factory $IncomeCreateCopyWith(IncomeCreate value, $Res Function(IncomeCreate) _then) = _$IncomeCreateCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode? regimeCode, String? description,@JsonKey(name: 'counterparty_type') CounterpartyType? counterpartyType,@JsonKey(name: 'amount_cents') NonNegativeCents? amountCents,@JsonKey(name: 'iva_rate') IvaRate? ivaRate,@JsonKey(name: 'iva_cents') NonNegativeCents? ivaCents,@JsonKey(name: 'isr_withheld_cents') NonNegativeCents? isrWithheldCents,@JsonKey(name: 'iva_withheld_cents') NonNegativeCents? ivaWithheldCents,@JsonKey(name: 'invoiced_on') DateTime? invoicedOn,@JsonKey(name: 'collected_on') DateTime? collectedOn, Uuid? id
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
@pragma('vm:prefer-inline') @override $Res call({Object? regimeCode = freezed,Object? description = freezed,Object? counterpartyType = freezed,Object? amountCents = freezed,Object? ivaRate = freezed,Object? ivaCents = freezed,Object? isrWithheldCents = freezed,Object? ivaWithheldCents = freezed,Object? invoicedOn = freezed,Object? collectedOn = freezed,Object? id = freezed,}) {
  return _then(IncomeCreate(
regimeCode: freezed == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode?,description: freezed == description ? _self.description : description // ignore: cast_nullable_to_non_nullable
as String?,counterpartyType: freezed == counterpartyType ? _self.counterpartyType : counterpartyType // ignore: cast_nullable_to_non_nullable
as CounterpartyType?,amountCents: freezed == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,ivaRate: freezed == ivaRate ? _self.ivaRate : ivaRate // ignore: cast_nullable_to_non_nullable
as IvaRate?,ivaCents: freezed == ivaCents ? _self.ivaCents : ivaCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,isrWithheldCents: freezed == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,ivaWithheldCents: freezed == ivaWithheldCents ? _self.ivaWithheldCents : ivaWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,invoicedOn: freezed == invoicedOn ? _self.invoicedOn : invoicedOn // ignore: cast_nullable_to_non_nullable
as DateTime?,collectedOn: freezed == collectedOn ? _self.collectedOn : collectedOn // ignore: cast_nullable_to_non_nullable
as DateTime?,id: freezed == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid?,
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode? regimeCode,  String? description, @JsonKey(name: 'counterparty_type')  CounterpartyType? counterpartyType, @JsonKey(name: 'amount_cents')  NonNegativeCents? amountCents, @JsonKey(name: 'iva_rate')  IvaRate? ivaRate, @JsonKey(name: 'iva_cents')  NonNegativeCents? ivaCents, @JsonKey(name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(name: 'iva_withheld_cents')  NonNegativeCents? ivaWithheldCents, @JsonKey(name: 'invoiced_on')  DateTime? invoicedOn, @JsonKey(name: 'collected_on')  DateTime? collectedOn,  Uuid? id)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _IncomeCreate() when $default != null:
return $default(_that.regimeCode,_that.description,_that.counterpartyType,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.isrWithheldCents,_that.ivaWithheldCents,_that.invoicedOn,_that.collectedOn,_that.id);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'regime_code')  RegimeCode? regimeCode,  String? description, @JsonKey(name: 'counterparty_type')  CounterpartyType? counterpartyType, @JsonKey(name: 'amount_cents')  NonNegativeCents? amountCents, @JsonKey(name: 'iva_rate')  IvaRate? ivaRate, @JsonKey(name: 'iva_cents')  NonNegativeCents? ivaCents, @JsonKey(name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(name: 'iva_withheld_cents')  NonNegativeCents? ivaWithheldCents, @JsonKey(name: 'invoiced_on')  DateTime? invoicedOn, @JsonKey(name: 'collected_on')  DateTime? collectedOn,  Uuid? id)  $default,) {final _that = this;
switch (_that) {
case _IncomeCreate():
return $default(_that.regimeCode,_that.description,_that.counterpartyType,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.isrWithheldCents,_that.ivaWithheldCents,_that.invoicedOn,_that.collectedOn,_that.id);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'regime_code')  RegimeCode? regimeCode,  String? description, @JsonKey(name: 'counterparty_type')  CounterpartyType? counterpartyType, @JsonKey(name: 'amount_cents')  NonNegativeCents? amountCents, @JsonKey(name: 'iva_rate')  IvaRate? ivaRate, @JsonKey(name: 'iva_cents')  NonNegativeCents? ivaCents, @JsonKey(name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(name: 'iva_withheld_cents')  NonNegativeCents? ivaWithheldCents, @JsonKey(name: 'invoiced_on')  DateTime? invoicedOn, @JsonKey(name: 'collected_on')  DateTime? collectedOn,  Uuid? id)?  $default,) {final _that = this;
switch (_that) {
case _IncomeCreate() when $default != null:
return $default(_that.regimeCode,_that.description,_that.counterpartyType,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.isrWithheldCents,_that.ivaWithheldCents,_that.invoicedOn,_that.collectedOn,_that.id);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _IncomeCreate implements IncomeCreate {
  const _IncomeCreate({@JsonKey(name: 'regime_code') this.regimeCode, this.description, @JsonKey(name: 'counterparty_type') this.counterpartyType, @JsonKey(name: 'amount_cents') this.amountCents, @JsonKey(name: 'iva_rate') this.ivaRate, @JsonKey(name: 'iva_cents') this.ivaCents, @JsonKey(name: 'isr_withheld_cents') this.isrWithheldCents, @JsonKey(name: 'iva_withheld_cents') this.ivaWithheldCents, @JsonKey(name: 'invoiced_on') this.invoicedOn, @JsonKey(name: 'collected_on') this.collectedOn, this.id});
  factory _IncomeCreate.fromJson(Map<String, dynamic> json) => _$IncomeCreateFromJson(json);

@override@JsonKey(name: 'regime_code') final  RegimeCode? regimeCode;
@override final  String? description;
@override@JsonKey(name: 'counterparty_type') final  CounterpartyType? counterpartyType;
/// Subtotal before IVA.
@override@JsonKey(name: 'amount_cents') final  NonNegativeCents? amountCents;
@override@JsonKey(name: 'iva_rate') final  IvaRate? ivaRate;
@override@JsonKey(name: 'iva_cents') final  NonNegativeCents? ivaCents;
@override@JsonKey(name: 'isr_withheld_cents') final  NonNegativeCents? isrWithheldCents;
@override@JsonKey(name: 'iva_withheld_cents') final  NonNegativeCents? ivaWithheldCents;
@override@JsonKey(name: 'invoiced_on') final  DateTime? invoicedOn;
/// Required for cash-basis regimes; may be null for PM 601 (pending collection).
@override@JsonKey(name: 'collected_on') final  DateTime? collectedOn;
@override final  Uuid? id;

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
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _IncomeCreate&&(identical(other.regimeCode, regimeCode) || other.regimeCode == regimeCode)&&(identical(other.description, description) || other.description == description)&&(identical(other.counterpartyType, counterpartyType) || other.counterpartyType == counterpartyType)&&(identical(other.amountCents, amountCents) || other.amountCents == amountCents)&&(identical(other.ivaRate, ivaRate) || other.ivaRate == ivaRate)&&(identical(other.ivaCents, ivaCents) || other.ivaCents == ivaCents)&&(identical(other.isrWithheldCents, isrWithheldCents) || other.isrWithheldCents == isrWithheldCents)&&(identical(other.ivaWithheldCents, ivaWithheldCents) || other.ivaWithheldCents == ivaWithheldCents)&&(identical(other.invoicedOn, invoicedOn) || other.invoicedOn == invoicedOn)&&(identical(other.collectedOn, collectedOn) || other.collectedOn == collectedOn)&&(identical(other.id, id) || other.id == id));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,regimeCode,description,counterpartyType,amountCents,ivaRate,ivaCents,isrWithheldCents,ivaWithheldCents,invoicedOn,collectedOn,id);
}

@override
String toString() {
    return 'IncomeCreate(regimeCode: $regimeCode, description: $description, counterpartyType: $counterpartyType, amountCents: $amountCents, ivaRate: $ivaRate, ivaCents: $ivaCents, isrWithheldCents: $isrWithheldCents, ivaWithheldCents: $ivaWithheldCents, invoicedOn: $invoicedOn, collectedOn: $collectedOn, id: $id)';
}


}

/// @nodoc
abstract mixin class _$IncomeCreateCopyWith<$Res> implements $IncomeCreateCopyWith<$Res> {
  factory _$IncomeCreateCopyWith(_IncomeCreate value, $Res Function(_IncomeCreate) _then) = __$IncomeCreateCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'regime_code') RegimeCode? regimeCode, String? description,@JsonKey(name: 'counterparty_type') CounterpartyType? counterpartyType,@JsonKey(name: 'amount_cents') NonNegativeCents? amountCents,@JsonKey(name: 'iva_rate') IvaRate? ivaRate,@JsonKey(name: 'iva_cents') NonNegativeCents? ivaCents,@JsonKey(name: 'isr_withheld_cents') NonNegativeCents? isrWithheldCents,@JsonKey(name: 'iva_withheld_cents') NonNegativeCents? ivaWithheldCents,@JsonKey(name: 'invoiced_on') DateTime? invoicedOn,@JsonKey(name: 'collected_on') DateTime? collectedOn, Uuid? id
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
@override @pragma('vm:prefer-inline') $Res call({Object? regimeCode = freezed,Object? description = freezed,Object? counterpartyType = freezed,Object? amountCents = freezed,Object? ivaRate = freezed,Object? ivaCents = freezed,Object? isrWithheldCents = freezed,Object? ivaWithheldCents = freezed,Object? invoicedOn = freezed,Object? collectedOn = freezed,Object? id = freezed,}) {
  return _then(_IncomeCreate(
regimeCode: freezed == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode?,description: freezed == description ? _self.description : description // ignore: cast_nullable_to_non_nullable
as String?,counterpartyType: freezed == counterpartyType ? _self.counterpartyType : counterpartyType // ignore: cast_nullable_to_non_nullable
as CounterpartyType?,amountCents: freezed == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,ivaRate: freezed == ivaRate ? _self.ivaRate : ivaRate // ignore: cast_nullable_to_non_nullable
as IvaRate?,ivaCents: freezed == ivaCents ? _self.ivaCents : ivaCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,isrWithheldCents: freezed == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,ivaWithheldCents: freezed == ivaWithheldCents ? _self.ivaWithheldCents : ivaWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,invoicedOn: freezed == invoicedOn ? _self.invoicedOn : invoicedOn // ignore: cast_nullable_to_non_nullable
as DateTime?,collectedOn: freezed == collectedOn ? _self.collectedOn : collectedOn // ignore: cast_nullable_to_non_nullable
as DateTime?,id: freezed == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid?,
  ));
}


}

// dart format on
