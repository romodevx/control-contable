// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'annual_credits.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$AnnualCredits {

@JsonKey(name: 'payments_cents') NonNegativeCents get paymentsCents;@JsonKey(name: 'withholdings_cents') NonNegativeCents get withholdingsCents;@JsonKey(name: 'salary_isr_cents') NonNegativeCents get salaryIsrCents;@JsonKey(name: 'total_cents') NonNegativeCents get totalCents;@JsonKey(name: 'payment_sources') PaymentSourceCounts get paymentSources;
/// Create a copy of AnnualCredits
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$AnnualCreditsCopyWith<AnnualCredits> get copyWith => _$AnnualCreditsCopyWithImpl<AnnualCredits>(this as AnnualCredits, _$identity);

  /// Serializes this AnnualCredits to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as AnnualCredits;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is AnnualCredits&&(identical(other.paymentsCents, _this.paymentsCents) || other.paymentsCents == _this.paymentsCents)&&(identical(other.withholdingsCents, _this.withholdingsCents) || other.withholdingsCents == _this.withholdingsCents)&&(identical(other.salaryIsrCents, _this.salaryIsrCents) || other.salaryIsrCents == _this.salaryIsrCents)&&(identical(other.totalCents, _this.totalCents) || other.totalCents == _this.totalCents)&&(identical(other.paymentSources, _this.paymentSources) || other.paymentSources == _this.paymentSources));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as AnnualCredits;
  return Object.hash(runtimeType,_this.paymentsCents,_this.withholdingsCents,_this.salaryIsrCents,_this.totalCents,_this.paymentSources);
}

@override
String toString() {
  final _this = this as AnnualCredits;
  return 'AnnualCredits(paymentsCents: ${_this.paymentsCents}, withholdingsCents: ${_this.withholdingsCents}, salaryIsrCents: ${_this.salaryIsrCents}, totalCents: ${_this.totalCents}, paymentSources: ${_this.paymentSources})';
}


}

/// @nodoc
abstract mixin class $AnnualCreditsCopyWith<$Res>  {
  factory $AnnualCreditsCopyWith(AnnualCredits value, $Res Function(AnnualCredits) _then) = _$AnnualCreditsCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'payments_cents') NonNegativeCents paymentsCents,@JsonKey(name: 'withholdings_cents') NonNegativeCents withholdingsCents,@JsonKey(name: 'salary_isr_cents') NonNegativeCents salaryIsrCents,@JsonKey(name: 'total_cents') NonNegativeCents totalCents,@JsonKey(name: 'payment_sources') PaymentSourceCounts paymentSources
});


$PaymentSourceCountsCopyWith<$Res> get paymentSources;

}
/// @nodoc
class _$AnnualCreditsCopyWithImpl<$Res>
    implements $AnnualCreditsCopyWith<$Res> {
  _$AnnualCreditsCopyWithImpl(this._self, this._then);

  final AnnualCredits _self;
  final $Res Function(AnnualCredits) _then;

/// Create a copy of AnnualCredits
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? paymentsCents = null,Object? withholdingsCents = null,Object? salaryIsrCents = null,Object? totalCents = null,Object? paymentSources = null,}) {
  return _then(AnnualCredits(
paymentsCents: null == paymentsCents ? _self.paymentsCents : paymentsCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,withholdingsCents: null == withholdingsCents ? _self.withholdingsCents : withholdingsCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,salaryIsrCents: null == salaryIsrCents ? _self.salaryIsrCents : salaryIsrCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,totalCents: null == totalCents ? _self.totalCents : totalCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,paymentSources: null == paymentSources ? _self.paymentSources : paymentSources // ignore: cast_nullable_to_non_nullable
as PaymentSourceCounts,
  ));
}
/// Create a copy of AnnualCredits
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$PaymentSourceCountsCopyWith<$Res> get paymentSources {
  
  return $PaymentSourceCountsCopyWith<$Res>(_self.paymentSources, (value) {
    return _then(_self.copyWith(paymentSources: value));
  });
}
}


/// Adds pattern-matching-related methods to [AnnualCredits].
extension AnnualCreditsPatterns on AnnualCredits {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _AnnualCredits value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _AnnualCredits() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _AnnualCredits value)  $default,){
final _that = this;
switch (_that) {
case _AnnualCredits():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _AnnualCredits value)?  $default,){
final _that = this;
switch (_that) {
case _AnnualCredits() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'payments_cents')  NonNegativeCents paymentsCents, @JsonKey(name: 'withholdings_cents')  NonNegativeCents withholdingsCents, @JsonKey(name: 'salary_isr_cents')  NonNegativeCents salaryIsrCents, @JsonKey(name: 'total_cents')  NonNegativeCents totalCents, @JsonKey(name: 'payment_sources')  PaymentSourceCounts paymentSources)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _AnnualCredits() when $default != null:
return $default(_that.paymentsCents,_that.withholdingsCents,_that.salaryIsrCents,_that.totalCents,_that.paymentSources);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'payments_cents')  NonNegativeCents paymentsCents, @JsonKey(name: 'withholdings_cents')  NonNegativeCents withholdingsCents, @JsonKey(name: 'salary_isr_cents')  NonNegativeCents salaryIsrCents, @JsonKey(name: 'total_cents')  NonNegativeCents totalCents, @JsonKey(name: 'payment_sources')  PaymentSourceCounts paymentSources)  $default,) {final _that = this;
switch (_that) {
case _AnnualCredits():
return $default(_that.paymentsCents,_that.withholdingsCents,_that.salaryIsrCents,_that.totalCents,_that.paymentSources);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'payments_cents')  NonNegativeCents paymentsCents, @JsonKey(name: 'withholdings_cents')  NonNegativeCents withholdingsCents, @JsonKey(name: 'salary_isr_cents')  NonNegativeCents salaryIsrCents, @JsonKey(name: 'total_cents')  NonNegativeCents totalCents, @JsonKey(name: 'payment_sources')  PaymentSourceCounts paymentSources)?  $default,) {final _that = this;
switch (_that) {
case _AnnualCredits() when $default != null:
return $default(_that.paymentsCents,_that.withholdingsCents,_that.salaryIsrCents,_that.totalCents,_that.paymentSources);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _AnnualCredits implements AnnualCredits {
  const _AnnualCredits({@JsonKey(name: 'payments_cents') required this.paymentsCents, @JsonKey(name: 'withholdings_cents') required this.withholdingsCents, @JsonKey(name: 'salary_isr_cents') required this.salaryIsrCents, @JsonKey(name: 'total_cents') required this.totalCents, @JsonKey(name: 'payment_sources') required this.paymentSources});
  factory _AnnualCredits.fromJson(Map<String, dynamic> json) => _$AnnualCreditsFromJson(json);

@override@JsonKey(name: 'payments_cents') final  NonNegativeCents paymentsCents;
@override@JsonKey(name: 'withholdings_cents') final  NonNegativeCents withholdingsCents;
@override@JsonKey(name: 'salary_isr_cents') final  NonNegativeCents salaryIsrCents;
@override@JsonKey(name: 'total_cents') final  NonNegativeCents totalCents;
@override@JsonKey(name: 'payment_sources') final  PaymentSourceCounts paymentSources;

/// Create a copy of AnnualCredits
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$AnnualCreditsCopyWith<_AnnualCredits> get copyWith => __$AnnualCreditsCopyWithImpl<_AnnualCredits>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$AnnualCreditsToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _AnnualCredits&&(identical(other.paymentsCents, paymentsCents) || other.paymentsCents == paymentsCents)&&(identical(other.withholdingsCents, withholdingsCents) || other.withholdingsCents == withholdingsCents)&&(identical(other.salaryIsrCents, salaryIsrCents) || other.salaryIsrCents == salaryIsrCents)&&(identical(other.totalCents, totalCents) || other.totalCents == totalCents)&&(identical(other.paymentSources, paymentSources) || other.paymentSources == paymentSources));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,paymentsCents,withholdingsCents,salaryIsrCents,totalCents,paymentSources);
}

@override
String toString() {
    return 'AnnualCredits(paymentsCents: $paymentsCents, withholdingsCents: $withholdingsCents, salaryIsrCents: $salaryIsrCents, totalCents: $totalCents, paymentSources: $paymentSources)';
}


}

/// @nodoc
abstract mixin class _$AnnualCreditsCopyWith<$Res> implements $AnnualCreditsCopyWith<$Res> {
  factory _$AnnualCreditsCopyWith(_AnnualCredits value, $Res Function(_AnnualCredits) _then) = __$AnnualCreditsCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'payments_cents') NonNegativeCents paymentsCents,@JsonKey(name: 'withholdings_cents') NonNegativeCents withholdingsCents,@JsonKey(name: 'salary_isr_cents') NonNegativeCents salaryIsrCents,@JsonKey(name: 'total_cents') NonNegativeCents totalCents,@JsonKey(name: 'payment_sources') PaymentSourceCounts paymentSources
});


@override $PaymentSourceCountsCopyWith<$Res> get paymentSources;

}
/// @nodoc
class __$AnnualCreditsCopyWithImpl<$Res>
    implements _$AnnualCreditsCopyWith<$Res> {
  __$AnnualCreditsCopyWithImpl(this._self, this._then);

  final _AnnualCredits _self;
  final $Res Function(_AnnualCredits) _then;

/// Create a copy of AnnualCredits
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? paymentsCents = null,Object? withholdingsCents = null,Object? salaryIsrCents = null,Object? totalCents = null,Object? paymentSources = null,}) {
  return _then(_AnnualCredits(
paymentsCents: null == paymentsCents ? _self.paymentsCents : paymentsCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,withholdingsCents: null == withholdingsCents ? _self.withholdingsCents : withholdingsCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,salaryIsrCents: null == salaryIsrCents ? _self.salaryIsrCents : salaryIsrCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,totalCents: null == totalCents ? _self.totalCents : totalCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,paymentSources: null == paymentSources ? _self.paymentSources : paymentSources // ignore: cast_nullable_to_non_nullable
as PaymentSourceCounts,
  ));
}

/// Create a copy of AnnualCredits
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$PaymentSourceCountsCopyWith<$Res> get paymentSources {
  
  return $PaymentSourceCountsCopyWith<$Res>(_self.paymentSources, (value) {
    return _then(_self.copyWith(paymentSources: value));
  });
}
}

// dart format on
