// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'annual_tax.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$AnnualTax {

@JsonKey(name: 'accumulable_income_cents') NonNegativeCents get accumulableIncomeCents;@JsonKey(includeIfNull: true, name: 'personal_deductions_cents') NonNegativeCents? get personalDeductionsCents;@JsonKey(name: 'base_cents') NonNegativeCents get baseCents;@JsonKey(name: 'isr_cents') NonNegativeCents get isrCents; AnnualCredits get credits; AnnualBalance get balance;
/// Create a copy of AnnualTax
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$AnnualTaxCopyWith<AnnualTax> get copyWith => _$AnnualTaxCopyWithImpl<AnnualTax>(this as AnnualTax, _$identity);

  /// Serializes this AnnualTax to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as AnnualTax;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is AnnualTax&&(identical(other.accumulableIncomeCents, _this.accumulableIncomeCents) || other.accumulableIncomeCents == _this.accumulableIncomeCents)&&(identical(other.personalDeductionsCents, _this.personalDeductionsCents) || other.personalDeductionsCents == _this.personalDeductionsCents)&&(identical(other.baseCents, _this.baseCents) || other.baseCents == _this.baseCents)&&(identical(other.isrCents, _this.isrCents) || other.isrCents == _this.isrCents)&&(identical(other.credits, _this.credits) || other.credits == _this.credits)&&(identical(other.balance, _this.balance) || other.balance == _this.balance));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as AnnualTax;
  return Object.hash(runtimeType,_this.accumulableIncomeCents,_this.personalDeductionsCents,_this.baseCents,_this.isrCents,_this.credits,_this.balance);
}

@override
String toString() {
  final _this = this as AnnualTax;
  return 'AnnualTax(accumulableIncomeCents: ${_this.accumulableIncomeCents}, personalDeductionsCents: ${_this.personalDeductionsCents}, baseCents: ${_this.baseCents}, isrCents: ${_this.isrCents}, credits: ${_this.credits}, balance: ${_this.balance})';
}


}

/// @nodoc
abstract mixin class $AnnualTaxCopyWith<$Res>  {
  factory $AnnualTaxCopyWith(AnnualTax value, $Res Function(AnnualTax) _then) = _$AnnualTaxCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'accumulable_income_cents') NonNegativeCents accumulableIncomeCents,@JsonKey(includeIfNull: true, name: 'personal_deductions_cents') NonNegativeCents? personalDeductionsCents,@JsonKey(name: 'base_cents') NonNegativeCents baseCents,@JsonKey(name: 'isr_cents') NonNegativeCents isrCents, AnnualCredits credits, AnnualBalance balance
});


$AnnualCreditsCopyWith<$Res> get credits;$AnnualBalanceCopyWith<$Res> get balance;

}
/// @nodoc
class _$AnnualTaxCopyWithImpl<$Res>
    implements $AnnualTaxCopyWith<$Res> {
  _$AnnualTaxCopyWithImpl(this._self, this._then);

  final AnnualTax _self;
  final $Res Function(AnnualTax) _then;

/// Create a copy of AnnualTax
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? accumulableIncomeCents = null,Object? personalDeductionsCents = freezed,Object? baseCents = null,Object? isrCents = null,Object? credits = null,Object? balance = null,}) {
  return _then(AnnualTax(
accumulableIncomeCents: null == accumulableIncomeCents ? _self.accumulableIncomeCents : accumulableIncomeCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,personalDeductionsCents: freezed == personalDeductionsCents ? _self.personalDeductionsCents : personalDeductionsCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,baseCents: null == baseCents ? _self.baseCents : baseCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,isrCents: null == isrCents ? _self.isrCents : isrCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,credits: null == credits ? _self.credits : credits // ignore: cast_nullable_to_non_nullable
as AnnualCredits,balance: null == balance ? _self.balance : balance // ignore: cast_nullable_to_non_nullable
as AnnualBalance,
  ));
}
/// Create a copy of AnnualTax
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$AnnualCreditsCopyWith<$Res> get credits {
  
  return $AnnualCreditsCopyWith<$Res>(_self.credits, (value) {
    return _then(_self.copyWith(credits: value));
  });
}/// Create a copy of AnnualTax
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$AnnualBalanceCopyWith<$Res> get balance {
  
  return $AnnualBalanceCopyWith<$Res>(_self.balance, (value) {
    return _then(_self.copyWith(balance: value));
  });
}
}


/// Adds pattern-matching-related methods to [AnnualTax].
extension AnnualTaxPatterns on AnnualTax {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _AnnualTax value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _AnnualTax() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _AnnualTax value)  $default,){
final _that = this;
switch (_that) {
case _AnnualTax():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _AnnualTax value)?  $default,){
final _that = this;
switch (_that) {
case _AnnualTax() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'accumulable_income_cents')  NonNegativeCents accumulableIncomeCents, @JsonKey(includeIfNull: true, name: 'personal_deductions_cents')  NonNegativeCents? personalDeductionsCents, @JsonKey(name: 'base_cents')  NonNegativeCents baseCents, @JsonKey(name: 'isr_cents')  NonNegativeCents isrCents,  AnnualCredits credits,  AnnualBalance balance)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _AnnualTax() when $default != null:
return $default(_that.accumulableIncomeCents,_that.personalDeductionsCents,_that.baseCents,_that.isrCents,_that.credits,_that.balance);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'accumulable_income_cents')  NonNegativeCents accumulableIncomeCents, @JsonKey(includeIfNull: true, name: 'personal_deductions_cents')  NonNegativeCents? personalDeductionsCents, @JsonKey(name: 'base_cents')  NonNegativeCents baseCents, @JsonKey(name: 'isr_cents')  NonNegativeCents isrCents,  AnnualCredits credits,  AnnualBalance balance)  $default,) {final _that = this;
switch (_that) {
case _AnnualTax():
return $default(_that.accumulableIncomeCents,_that.personalDeductionsCents,_that.baseCents,_that.isrCents,_that.credits,_that.balance);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'accumulable_income_cents')  NonNegativeCents accumulableIncomeCents, @JsonKey(includeIfNull: true, name: 'personal_deductions_cents')  NonNegativeCents? personalDeductionsCents, @JsonKey(name: 'base_cents')  NonNegativeCents baseCents, @JsonKey(name: 'isr_cents')  NonNegativeCents isrCents,  AnnualCredits credits,  AnnualBalance balance)?  $default,) {final _that = this;
switch (_that) {
case _AnnualTax() when $default != null:
return $default(_that.accumulableIncomeCents,_that.personalDeductionsCents,_that.baseCents,_that.isrCents,_that.credits,_that.balance);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _AnnualTax implements AnnualTax {
  const _AnnualTax({@JsonKey(name: 'accumulable_income_cents') required this.accumulableIncomeCents, @JsonKey(includeIfNull: true, name: 'personal_deductions_cents') required this.personalDeductionsCents, @JsonKey(name: 'base_cents') required this.baseCents, @JsonKey(name: 'isr_cents') required this.isrCents, required this.credits, required this.balance});
  factory _AnnualTax.fromJson(Map<String, dynamic> json) => _$AnnualTaxFromJson(json);

@override@JsonKey(name: 'accumulable_income_cents') final  NonNegativeCents accumulableIncomeCents;
@override@JsonKey(includeIfNull: true, name: 'personal_deductions_cents') final  NonNegativeCents? personalDeductionsCents;
@override@JsonKey(name: 'base_cents') final  NonNegativeCents baseCents;
@override@JsonKey(name: 'isr_cents') final  NonNegativeCents isrCents;
@override final  AnnualCredits credits;
@override final  AnnualBalance balance;

/// Create a copy of AnnualTax
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$AnnualTaxCopyWith<_AnnualTax> get copyWith => __$AnnualTaxCopyWithImpl<_AnnualTax>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$AnnualTaxToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _AnnualTax&&(identical(other.accumulableIncomeCents, accumulableIncomeCents) || other.accumulableIncomeCents == accumulableIncomeCents)&&(identical(other.personalDeductionsCents, personalDeductionsCents) || other.personalDeductionsCents == personalDeductionsCents)&&(identical(other.baseCents, baseCents) || other.baseCents == baseCents)&&(identical(other.isrCents, isrCents) || other.isrCents == isrCents)&&(identical(other.credits, credits) || other.credits == credits)&&(identical(other.balance, balance) || other.balance == balance));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,accumulableIncomeCents,personalDeductionsCents,baseCents,isrCents,credits,balance);
}

@override
String toString() {
    return 'AnnualTax(accumulableIncomeCents: $accumulableIncomeCents, personalDeductionsCents: $personalDeductionsCents, baseCents: $baseCents, isrCents: $isrCents, credits: $credits, balance: $balance)';
}


}

/// @nodoc
abstract mixin class _$AnnualTaxCopyWith<$Res> implements $AnnualTaxCopyWith<$Res> {
  factory _$AnnualTaxCopyWith(_AnnualTax value, $Res Function(_AnnualTax) _then) = __$AnnualTaxCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'accumulable_income_cents') NonNegativeCents accumulableIncomeCents,@JsonKey(includeIfNull: true, name: 'personal_deductions_cents') NonNegativeCents? personalDeductionsCents,@JsonKey(name: 'base_cents') NonNegativeCents baseCents,@JsonKey(name: 'isr_cents') NonNegativeCents isrCents, AnnualCredits credits, AnnualBalance balance
});


@override $AnnualCreditsCopyWith<$Res> get credits;@override $AnnualBalanceCopyWith<$Res> get balance;

}
/// @nodoc
class __$AnnualTaxCopyWithImpl<$Res>
    implements _$AnnualTaxCopyWith<$Res> {
  __$AnnualTaxCopyWithImpl(this._self, this._then);

  final _AnnualTax _self;
  final $Res Function(_AnnualTax) _then;

/// Create a copy of AnnualTax
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? accumulableIncomeCents = null,Object? personalDeductionsCents = freezed,Object? baseCents = null,Object? isrCents = null,Object? credits = null,Object? balance = null,}) {
  return _then(_AnnualTax(
accumulableIncomeCents: null == accumulableIncomeCents ? _self.accumulableIncomeCents : accumulableIncomeCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,personalDeductionsCents: freezed == personalDeductionsCents ? _self.personalDeductionsCents : personalDeductionsCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,baseCents: null == baseCents ? _self.baseCents : baseCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,isrCents: null == isrCents ? _self.isrCents : isrCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,credits: null == credits ? _self.credits : credits // ignore: cast_nullable_to_non_nullable
as AnnualCredits,balance: null == balance ? _self.balance : balance // ignore: cast_nullable_to_non_nullable
as AnnualBalance,
  ));
}

/// Create a copy of AnnualTax
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$AnnualCreditsCopyWith<$Res> get credits {
  
  return $AnnualCreditsCopyWith<$Res>(_self.credits, (value) {
    return _then(_self.copyWith(credits: value));
  });
}/// Create a copy of AnnualTax
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$AnnualBalanceCopyWith<$Res> get balance {
  
  return $AnnualBalanceCopyWith<$Res>(_self.balance, (value) {
    return _then(_self.copyWith(balance: value));
  });
}
}

// dart format on
