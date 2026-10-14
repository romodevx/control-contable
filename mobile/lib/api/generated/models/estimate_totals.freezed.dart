// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'estimate_totals.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$EstimateTotals {

/// Monthly IVA consolidated across the taxpayer's regimes.
 TaxResult get iva;/// Sum of the payable per-regime provisional ISR; favor balances stay per regime (OQ-T10).
 TaxResult get isr;/// Withheld ISR/IVA to remit in full, shown on its own line (never netted against IVA, LIVA 1-A).
@JsonKey(name: 'withholdings_to_remit_cents') NonNegativeCents get withholdingsToRemitCents;@JsonKey(name: 'total_payable_rounded_cents') NonNegativeCents get totalPayableRoundedCents;
/// Create a copy of EstimateTotals
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$EstimateTotalsCopyWith<EstimateTotals> get copyWith => _$EstimateTotalsCopyWithImpl<EstimateTotals>(this as EstimateTotals, _$identity);

  /// Serializes this EstimateTotals to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as EstimateTotals;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is EstimateTotals&&(identical(other.iva, _this.iva) || other.iva == _this.iva)&&(identical(other.isr, _this.isr) || other.isr == _this.isr)&&(identical(other.withholdingsToRemitCents, _this.withholdingsToRemitCents) || other.withholdingsToRemitCents == _this.withholdingsToRemitCents)&&(identical(other.totalPayableRoundedCents, _this.totalPayableRoundedCents) || other.totalPayableRoundedCents == _this.totalPayableRoundedCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as EstimateTotals;
  return Object.hash(runtimeType,_this.iva,_this.isr,_this.withholdingsToRemitCents,_this.totalPayableRoundedCents);
}

@override
String toString() {
  final _this = this as EstimateTotals;
  return 'EstimateTotals(iva: ${_this.iva}, isr: ${_this.isr}, withholdingsToRemitCents: ${_this.withholdingsToRemitCents}, totalPayableRoundedCents: ${_this.totalPayableRoundedCents})';
}


}

/// @nodoc
abstract mixin class $EstimateTotalsCopyWith<$Res>  {
  factory $EstimateTotalsCopyWith(EstimateTotals value, $Res Function(EstimateTotals) _then) = _$EstimateTotalsCopyWithImpl;
@useResult
$Res call({
 TaxResult iva, TaxResult isr,@JsonKey(name: 'withholdings_to_remit_cents') NonNegativeCents withholdingsToRemitCents,@JsonKey(name: 'total_payable_rounded_cents') NonNegativeCents totalPayableRoundedCents
});


$TaxResultCopyWith<$Res> get iva;$TaxResultCopyWith<$Res> get isr;

}
/// @nodoc
class _$EstimateTotalsCopyWithImpl<$Res>
    implements $EstimateTotalsCopyWith<$Res> {
  _$EstimateTotalsCopyWithImpl(this._self, this._then);

  final EstimateTotals _self;
  final $Res Function(EstimateTotals) _then;

/// Create a copy of EstimateTotals
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? iva = null,Object? isr = null,Object? withholdingsToRemitCents = null,Object? totalPayableRoundedCents = null,}) {
  return _then(EstimateTotals(
iva: null == iva ? _self.iva : iva // ignore: cast_nullable_to_non_nullable
as TaxResult,isr: null == isr ? _self.isr : isr // ignore: cast_nullable_to_non_nullable
as TaxResult,withholdingsToRemitCents: null == withholdingsToRemitCents ? _self.withholdingsToRemitCents : withholdingsToRemitCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,totalPayableRoundedCents: null == totalPayableRoundedCents ? _self.totalPayableRoundedCents : totalPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,
  ));
}
/// Create a copy of EstimateTotals
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TaxResultCopyWith<$Res> get iva {
  
  return $TaxResultCopyWith<$Res>(_self.iva, (value) {
    return _then(_self.copyWith(iva: value));
  });
}/// Create a copy of EstimateTotals
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TaxResultCopyWith<$Res> get isr {
  
  return $TaxResultCopyWith<$Res>(_self.isr, (value) {
    return _then(_self.copyWith(isr: value));
  });
}
}


/// Adds pattern-matching-related methods to [EstimateTotals].
extension EstimateTotalsPatterns on EstimateTotals {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _EstimateTotals value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _EstimateTotals() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _EstimateTotals value)  $default,){
final _that = this;
switch (_that) {
case _EstimateTotals():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _EstimateTotals value)?  $default,){
final _that = this;
switch (_that) {
case _EstimateTotals() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( TaxResult iva,  TaxResult isr, @JsonKey(name: 'withholdings_to_remit_cents')  NonNegativeCents withholdingsToRemitCents, @JsonKey(name: 'total_payable_rounded_cents')  NonNegativeCents totalPayableRoundedCents)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _EstimateTotals() when $default != null:
return $default(_that.iva,_that.isr,_that.withholdingsToRemitCents,_that.totalPayableRoundedCents);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( TaxResult iva,  TaxResult isr, @JsonKey(name: 'withholdings_to_remit_cents')  NonNegativeCents withholdingsToRemitCents, @JsonKey(name: 'total_payable_rounded_cents')  NonNegativeCents totalPayableRoundedCents)  $default,) {final _that = this;
switch (_that) {
case _EstimateTotals():
return $default(_that.iva,_that.isr,_that.withholdingsToRemitCents,_that.totalPayableRoundedCents);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( TaxResult iva,  TaxResult isr, @JsonKey(name: 'withholdings_to_remit_cents')  NonNegativeCents withholdingsToRemitCents, @JsonKey(name: 'total_payable_rounded_cents')  NonNegativeCents totalPayableRoundedCents)?  $default,) {final _that = this;
switch (_that) {
case _EstimateTotals() when $default != null:
return $default(_that.iva,_that.isr,_that.withholdingsToRemitCents,_that.totalPayableRoundedCents);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _EstimateTotals implements EstimateTotals {
  const _EstimateTotals({required this.iva, required this.isr, @JsonKey(name: 'withholdings_to_remit_cents') required this.withholdingsToRemitCents, @JsonKey(name: 'total_payable_rounded_cents') required this.totalPayableRoundedCents});
  factory _EstimateTotals.fromJson(Map<String, dynamic> json) => _$EstimateTotalsFromJson(json);

/// Monthly IVA consolidated across the taxpayer's regimes.
@override final  TaxResult iva;
/// Sum of the payable per-regime provisional ISR; favor balances stay per regime (OQ-T10).
@override final  TaxResult isr;
/// Withheld ISR/IVA to remit in full, shown on its own line (never netted against IVA, LIVA 1-A).
@override@JsonKey(name: 'withholdings_to_remit_cents') final  NonNegativeCents withholdingsToRemitCents;
@override@JsonKey(name: 'total_payable_rounded_cents') final  NonNegativeCents totalPayableRoundedCents;

/// Create a copy of EstimateTotals
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$EstimateTotalsCopyWith<_EstimateTotals> get copyWith => __$EstimateTotalsCopyWithImpl<_EstimateTotals>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$EstimateTotalsToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _EstimateTotals&&(identical(other.iva, iva) || other.iva == iva)&&(identical(other.isr, isr) || other.isr == isr)&&(identical(other.withholdingsToRemitCents, withholdingsToRemitCents) || other.withholdingsToRemitCents == withholdingsToRemitCents)&&(identical(other.totalPayableRoundedCents, totalPayableRoundedCents) || other.totalPayableRoundedCents == totalPayableRoundedCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,iva,isr,withholdingsToRemitCents,totalPayableRoundedCents);
}

@override
String toString() {
    return 'EstimateTotals(iva: $iva, isr: $isr, withholdingsToRemitCents: $withholdingsToRemitCents, totalPayableRoundedCents: $totalPayableRoundedCents)';
}


}

/// @nodoc
abstract mixin class _$EstimateTotalsCopyWith<$Res> implements $EstimateTotalsCopyWith<$Res> {
  factory _$EstimateTotalsCopyWith(_EstimateTotals value, $Res Function(_EstimateTotals) _then) = __$EstimateTotalsCopyWithImpl;
@override @useResult
$Res call({
 TaxResult iva, TaxResult isr,@JsonKey(name: 'withholdings_to_remit_cents') NonNegativeCents withholdingsToRemitCents,@JsonKey(name: 'total_payable_rounded_cents') NonNegativeCents totalPayableRoundedCents
});


@override $TaxResultCopyWith<$Res> get iva;@override $TaxResultCopyWith<$Res> get isr;

}
/// @nodoc
class __$EstimateTotalsCopyWithImpl<$Res>
    implements _$EstimateTotalsCopyWith<$Res> {
  __$EstimateTotalsCopyWithImpl(this._self, this._then);

  final _EstimateTotals _self;
  final $Res Function(_EstimateTotals) _then;

/// Create a copy of EstimateTotals
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? iva = null,Object? isr = null,Object? withholdingsToRemitCents = null,Object? totalPayableRoundedCents = null,}) {
  return _then(_EstimateTotals(
iva: null == iva ? _self.iva : iva // ignore: cast_nullable_to_non_nullable
as TaxResult,isr: null == isr ? _self.isr : isr // ignore: cast_nullable_to_non_nullable
as TaxResult,withholdingsToRemitCents: null == withholdingsToRemitCents ? _self.withholdingsToRemitCents : withholdingsToRemitCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,totalPayableRoundedCents: null == totalPayableRoundedCents ? _self.totalPayableRoundedCents : totalPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,
  ));
}

/// Create a copy of EstimateTotals
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TaxResultCopyWith<$Res> get iva {
  
  return $TaxResultCopyWith<$Res>(_self.iva, (value) {
    return _then(_self.copyWith(iva: value));
  });
}/// Create a copy of EstimateTotals
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$TaxResultCopyWith<$Res> get isr {
  
  return $TaxResultCopyWith<$Res>(_self.isr, (value) {
    return _then(_self.copyWith(isr: value));
  });
}
}

// dart format on
