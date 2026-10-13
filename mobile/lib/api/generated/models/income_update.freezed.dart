// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'income_update.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$IncomeUpdate {

@JsonKey(includeIfNull: false, name: 'regime_code') RegimeCode? get regimeCode;@JsonKey(includeIfNull: false) String? get description;@JsonKey(includeIfNull: false, name: 'counterparty_type') CounterpartyType? get counterpartyType;/// Subtotal before IVA.
@JsonKey(includeIfNull: false, name: 'amount_cents') NonNegativeCents? get amountCents;@JsonKey(includeIfNull: false, name: 'iva_rate') IvaRate? get ivaRate;@JsonKey(includeIfNull: false, name: 'iva_cents') NonNegativeCents? get ivaCents;@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') NonNegativeCents? get isrWithheldCents;@JsonKey(includeIfNull: false, name: 'iva_withheld_cents') NonNegativeCents? get ivaWithheldCents;@JsonKey(includeIfNull: false, name: 'invoiced_on') DateTime? get invoicedOn;/// Required for cash-basis regimes; may be null for PM 601 (pending collection).
@JsonKey(includeIfNull: false, name: 'collected_on') DateTime? get collectedOn;
/// Create a copy of IncomeUpdate
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$IncomeUpdateCopyWith<IncomeUpdate> get copyWith => _$IncomeUpdateCopyWithImpl<IncomeUpdate>(this as IncomeUpdate, _$identity);

  /// Serializes this IncomeUpdate to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as IncomeUpdate;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is IncomeUpdate&&(identical(other.regimeCode, _this.regimeCode) || other.regimeCode == _this.regimeCode)&&(identical(other.description, _this.description) || other.description == _this.description)&&(identical(other.counterpartyType, _this.counterpartyType) || other.counterpartyType == _this.counterpartyType)&&(identical(other.amountCents, _this.amountCents) || other.amountCents == _this.amountCents)&&(identical(other.ivaRate, _this.ivaRate) || other.ivaRate == _this.ivaRate)&&(identical(other.ivaCents, _this.ivaCents) || other.ivaCents == _this.ivaCents)&&(identical(other.isrWithheldCents, _this.isrWithheldCents) || other.isrWithheldCents == _this.isrWithheldCents)&&(identical(other.ivaWithheldCents, _this.ivaWithheldCents) || other.ivaWithheldCents == _this.ivaWithheldCents)&&(identical(other.invoicedOn, _this.invoicedOn) || other.invoicedOn == _this.invoicedOn)&&(identical(other.collectedOn, _this.collectedOn) || other.collectedOn == _this.collectedOn));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as IncomeUpdate;
  return Object.hash(runtimeType,_this.regimeCode,_this.description,_this.counterpartyType,_this.amountCents,_this.ivaRate,_this.ivaCents,_this.isrWithheldCents,_this.ivaWithheldCents,_this.invoicedOn,_this.collectedOn);
}

@override
String toString() {
  final _this = this as IncomeUpdate;
  return 'IncomeUpdate(regimeCode: ${_this.regimeCode}, description: ${_this.description}, counterpartyType: ${_this.counterpartyType}, amountCents: ${_this.amountCents}, ivaRate: ${_this.ivaRate}, ivaCents: ${_this.ivaCents}, isrWithheldCents: ${_this.isrWithheldCents}, ivaWithheldCents: ${_this.ivaWithheldCents}, invoicedOn: ${_this.invoicedOn}, collectedOn: ${_this.collectedOn})';
}


}

/// @nodoc
abstract mixin class $IncomeUpdateCopyWith<$Res>  {
  factory $IncomeUpdateCopyWith(IncomeUpdate value, $Res Function(IncomeUpdate) _then) = _$IncomeUpdateCopyWithImpl;
@useResult
$Res call({
@JsonKey(includeIfNull: false, name: 'regime_code') RegimeCode? regimeCode,@JsonKey(includeIfNull: false) String? description,@JsonKey(includeIfNull: false, name: 'counterparty_type') CounterpartyType? counterpartyType,@JsonKey(includeIfNull: false, name: 'amount_cents') NonNegativeCents? amountCents,@JsonKey(includeIfNull: false, name: 'iva_rate') IvaRate? ivaRate,@JsonKey(includeIfNull: false, name: 'iva_cents') NonNegativeCents? ivaCents,@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') NonNegativeCents? isrWithheldCents,@JsonKey(includeIfNull: false, name: 'iva_withheld_cents') NonNegativeCents? ivaWithheldCents,@JsonKey(includeIfNull: false, name: 'invoiced_on') DateTime? invoicedOn,@JsonKey(includeIfNull: false, name: 'collected_on') DateTime? collectedOn
});




}
/// @nodoc
class _$IncomeUpdateCopyWithImpl<$Res>
    implements $IncomeUpdateCopyWith<$Res> {
  _$IncomeUpdateCopyWithImpl(this._self, this._then);

  final IncomeUpdate _self;
  final $Res Function(IncomeUpdate) _then;

/// Create a copy of IncomeUpdate
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? regimeCode = freezed,Object? description = freezed,Object? counterpartyType = freezed,Object? amountCents = freezed,Object? ivaRate = freezed,Object? ivaCents = freezed,Object? isrWithheldCents = freezed,Object? ivaWithheldCents = freezed,Object? invoicedOn = freezed,Object? collectedOn = freezed,}) {
  return _then(IncomeUpdate(
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
as DateTime?,
  ));
}

}


/// Adds pattern-matching-related methods to [IncomeUpdate].
extension IncomeUpdatePatterns on IncomeUpdate {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _IncomeUpdate value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _IncomeUpdate() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _IncomeUpdate value)  $default,){
final _that = this;
switch (_that) {
case _IncomeUpdate():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _IncomeUpdate value)?  $default,){
final _that = this;
switch (_that) {
case _IncomeUpdate() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(includeIfNull: false, name: 'regime_code')  RegimeCode? regimeCode, @JsonKey(includeIfNull: false)  String? description, @JsonKey(includeIfNull: false, name: 'counterparty_type')  CounterpartyType? counterpartyType, @JsonKey(includeIfNull: false, name: 'amount_cents')  NonNegativeCents? amountCents, @JsonKey(includeIfNull: false, name: 'iva_rate')  IvaRate? ivaRate, @JsonKey(includeIfNull: false, name: 'iva_cents')  NonNegativeCents? ivaCents, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(includeIfNull: false, name: 'iva_withheld_cents')  NonNegativeCents? ivaWithheldCents, @JsonKey(includeIfNull: false, name: 'invoiced_on')  DateTime? invoicedOn, @JsonKey(includeIfNull: false, name: 'collected_on')  DateTime? collectedOn)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _IncomeUpdate() when $default != null:
return $default(_that.regimeCode,_that.description,_that.counterpartyType,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.isrWithheldCents,_that.ivaWithheldCents,_that.invoicedOn,_that.collectedOn);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(includeIfNull: false, name: 'regime_code')  RegimeCode? regimeCode, @JsonKey(includeIfNull: false)  String? description, @JsonKey(includeIfNull: false, name: 'counterparty_type')  CounterpartyType? counterpartyType, @JsonKey(includeIfNull: false, name: 'amount_cents')  NonNegativeCents? amountCents, @JsonKey(includeIfNull: false, name: 'iva_rate')  IvaRate? ivaRate, @JsonKey(includeIfNull: false, name: 'iva_cents')  NonNegativeCents? ivaCents, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(includeIfNull: false, name: 'iva_withheld_cents')  NonNegativeCents? ivaWithheldCents, @JsonKey(includeIfNull: false, name: 'invoiced_on')  DateTime? invoicedOn, @JsonKey(includeIfNull: false, name: 'collected_on')  DateTime? collectedOn)  $default,) {final _that = this;
switch (_that) {
case _IncomeUpdate():
return $default(_that.regimeCode,_that.description,_that.counterpartyType,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.isrWithheldCents,_that.ivaWithheldCents,_that.invoicedOn,_that.collectedOn);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(includeIfNull: false, name: 'regime_code')  RegimeCode? regimeCode, @JsonKey(includeIfNull: false)  String? description, @JsonKey(includeIfNull: false, name: 'counterparty_type')  CounterpartyType? counterpartyType, @JsonKey(includeIfNull: false, name: 'amount_cents')  NonNegativeCents? amountCents, @JsonKey(includeIfNull: false, name: 'iva_rate')  IvaRate? ivaRate, @JsonKey(includeIfNull: false, name: 'iva_cents')  NonNegativeCents? ivaCents, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(includeIfNull: false, name: 'iva_withheld_cents')  NonNegativeCents? ivaWithheldCents, @JsonKey(includeIfNull: false, name: 'invoiced_on')  DateTime? invoicedOn, @JsonKey(includeIfNull: false, name: 'collected_on')  DateTime? collectedOn)?  $default,) {final _that = this;
switch (_that) {
case _IncomeUpdate() when $default != null:
return $default(_that.regimeCode,_that.description,_that.counterpartyType,_that.amountCents,_that.ivaRate,_that.ivaCents,_that.isrWithheldCents,_that.ivaWithheldCents,_that.invoicedOn,_that.collectedOn);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _IncomeUpdate implements IncomeUpdate {
  const _IncomeUpdate({@JsonKey(includeIfNull: false, name: 'regime_code') this.regimeCode, @JsonKey(includeIfNull: false) this.description, @JsonKey(includeIfNull: false, name: 'counterparty_type') this.counterpartyType, @JsonKey(includeIfNull: false, name: 'amount_cents') this.amountCents, @JsonKey(includeIfNull: false, name: 'iva_rate') this.ivaRate, @JsonKey(includeIfNull: false, name: 'iva_cents') this.ivaCents, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents') this.isrWithheldCents, @JsonKey(includeIfNull: false, name: 'iva_withheld_cents') this.ivaWithheldCents, @JsonKey(includeIfNull: false, name: 'invoiced_on') this.invoicedOn, @JsonKey(includeIfNull: false, name: 'collected_on') this.collectedOn});
  factory _IncomeUpdate.fromJson(Map<String, dynamic> json) => _$IncomeUpdateFromJson(json);

@override@JsonKey(includeIfNull: false, name: 'regime_code') final  RegimeCode? regimeCode;
@override@JsonKey(includeIfNull: false) final  String? description;
@override@JsonKey(includeIfNull: false, name: 'counterparty_type') final  CounterpartyType? counterpartyType;
/// Subtotal before IVA.
@override@JsonKey(includeIfNull: false, name: 'amount_cents') final  NonNegativeCents? amountCents;
@override@JsonKey(includeIfNull: false, name: 'iva_rate') final  IvaRate? ivaRate;
@override@JsonKey(includeIfNull: false, name: 'iva_cents') final  NonNegativeCents? ivaCents;
@override@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') final  NonNegativeCents? isrWithheldCents;
@override@JsonKey(includeIfNull: false, name: 'iva_withheld_cents') final  NonNegativeCents? ivaWithheldCents;
@override@JsonKey(includeIfNull: false, name: 'invoiced_on') final  DateTime? invoicedOn;
/// Required for cash-basis regimes; may be null for PM 601 (pending collection).
@override@JsonKey(includeIfNull: false, name: 'collected_on') final  DateTime? collectedOn;

/// Create a copy of IncomeUpdate
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$IncomeUpdateCopyWith<_IncomeUpdate> get copyWith => __$IncomeUpdateCopyWithImpl<_IncomeUpdate>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$IncomeUpdateToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _IncomeUpdate&&(identical(other.regimeCode, regimeCode) || other.regimeCode == regimeCode)&&(identical(other.description, description) || other.description == description)&&(identical(other.counterpartyType, counterpartyType) || other.counterpartyType == counterpartyType)&&(identical(other.amountCents, amountCents) || other.amountCents == amountCents)&&(identical(other.ivaRate, ivaRate) || other.ivaRate == ivaRate)&&(identical(other.ivaCents, ivaCents) || other.ivaCents == ivaCents)&&(identical(other.isrWithheldCents, isrWithheldCents) || other.isrWithheldCents == isrWithheldCents)&&(identical(other.ivaWithheldCents, ivaWithheldCents) || other.ivaWithheldCents == ivaWithheldCents)&&(identical(other.invoicedOn, invoicedOn) || other.invoicedOn == invoicedOn)&&(identical(other.collectedOn, collectedOn) || other.collectedOn == collectedOn));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,regimeCode,description,counterpartyType,amountCents,ivaRate,ivaCents,isrWithheldCents,ivaWithheldCents,invoicedOn,collectedOn);
}

@override
String toString() {
    return 'IncomeUpdate(regimeCode: $regimeCode, description: $description, counterpartyType: $counterpartyType, amountCents: $amountCents, ivaRate: $ivaRate, ivaCents: $ivaCents, isrWithheldCents: $isrWithheldCents, ivaWithheldCents: $ivaWithheldCents, invoicedOn: $invoicedOn, collectedOn: $collectedOn)';
}


}

/// @nodoc
abstract mixin class _$IncomeUpdateCopyWith<$Res> implements $IncomeUpdateCopyWith<$Res> {
  factory _$IncomeUpdateCopyWith(_IncomeUpdate value, $Res Function(_IncomeUpdate) _then) = __$IncomeUpdateCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(includeIfNull: false, name: 'regime_code') RegimeCode? regimeCode,@JsonKey(includeIfNull: false) String? description,@JsonKey(includeIfNull: false, name: 'counterparty_type') CounterpartyType? counterpartyType,@JsonKey(includeIfNull: false, name: 'amount_cents') NonNegativeCents? amountCents,@JsonKey(includeIfNull: false, name: 'iva_rate') IvaRate? ivaRate,@JsonKey(includeIfNull: false, name: 'iva_cents') NonNegativeCents? ivaCents,@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') NonNegativeCents? isrWithheldCents,@JsonKey(includeIfNull: false, name: 'iva_withheld_cents') NonNegativeCents? ivaWithheldCents,@JsonKey(includeIfNull: false, name: 'invoiced_on') DateTime? invoicedOn,@JsonKey(includeIfNull: false, name: 'collected_on') DateTime? collectedOn
});




}
/// @nodoc
class __$IncomeUpdateCopyWithImpl<$Res>
    implements _$IncomeUpdateCopyWith<$Res> {
  __$IncomeUpdateCopyWithImpl(this._self, this._then);

  final _IncomeUpdate _self;
  final $Res Function(_IncomeUpdate) _then;

/// Create a copy of IncomeUpdate
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? regimeCode = freezed,Object? description = freezed,Object? counterpartyType = freezed,Object? amountCents = freezed,Object? ivaRate = freezed,Object? ivaCents = freezed,Object? isrWithheldCents = freezed,Object? ivaWithheldCents = freezed,Object? invoicedOn = freezed,Object? collectedOn = freezed,}) {
  return _then(_IncomeUpdate(
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
as DateTime?,
  ));
}


}

// dart format on
