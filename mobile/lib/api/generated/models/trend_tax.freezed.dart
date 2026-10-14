// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'trend_tax.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$TrendTax {

@JsonKey(name: 'iva_payable_rounded_cents') NonNegativeCents get ivaPayableRoundedCents;@JsonKey(name: 'isr_payable_rounded_cents') NonNegativeCents get isrPayableRoundedCents;@JsonKey(name: 'withholdings_to_remit_cents') NonNegativeCents get withholdingsToRemitCents;@JsonKey(name: 'total_payable_rounded_cents') NonNegativeCents get totalPayableRoundedCents;@JsonKey(name: 'iva_outcome') TrendTaxIvaOutcome get ivaOutcome; TaxpayerEstimateStatus get status;
/// Create a copy of TrendTax
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TrendTaxCopyWith<TrendTax> get copyWith => _$TrendTaxCopyWithImpl<TrendTax>(this as TrendTax, _$identity);

  /// Serializes this TrendTax to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as TrendTax;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TrendTax&&(identical(other.ivaPayableRoundedCents, _this.ivaPayableRoundedCents) || other.ivaPayableRoundedCents == _this.ivaPayableRoundedCents)&&(identical(other.isrPayableRoundedCents, _this.isrPayableRoundedCents) || other.isrPayableRoundedCents == _this.isrPayableRoundedCents)&&(identical(other.withholdingsToRemitCents, _this.withholdingsToRemitCents) || other.withholdingsToRemitCents == _this.withholdingsToRemitCents)&&(identical(other.totalPayableRoundedCents, _this.totalPayableRoundedCents) || other.totalPayableRoundedCents == _this.totalPayableRoundedCents)&&(identical(other.ivaOutcome, _this.ivaOutcome) || other.ivaOutcome == _this.ivaOutcome)&&(identical(other.status, _this.status) || other.status == _this.status));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TrendTax;
  return Object.hash(runtimeType,_this.ivaPayableRoundedCents,_this.isrPayableRoundedCents,_this.withholdingsToRemitCents,_this.totalPayableRoundedCents,_this.ivaOutcome,_this.status);
}

@override
String toString() {
  final _this = this as TrendTax;
  return 'TrendTax(ivaPayableRoundedCents: ${_this.ivaPayableRoundedCents}, isrPayableRoundedCents: ${_this.isrPayableRoundedCents}, withholdingsToRemitCents: ${_this.withholdingsToRemitCents}, totalPayableRoundedCents: ${_this.totalPayableRoundedCents}, ivaOutcome: ${_this.ivaOutcome}, status: ${_this.status})';
}


}

/// @nodoc
abstract mixin class $TrendTaxCopyWith<$Res>  {
  factory $TrendTaxCopyWith(TrendTax value, $Res Function(TrendTax) _then) = _$TrendTaxCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'iva_payable_rounded_cents') NonNegativeCents ivaPayableRoundedCents,@JsonKey(name: 'isr_payable_rounded_cents') NonNegativeCents isrPayableRoundedCents,@JsonKey(name: 'withholdings_to_remit_cents') NonNegativeCents withholdingsToRemitCents,@JsonKey(name: 'total_payable_rounded_cents') NonNegativeCents totalPayableRoundedCents,@JsonKey(name: 'iva_outcome') TrendTaxIvaOutcome ivaOutcome, TaxpayerEstimateStatus status
});




}
/// @nodoc
class _$TrendTaxCopyWithImpl<$Res>
    implements $TrendTaxCopyWith<$Res> {
  _$TrendTaxCopyWithImpl(this._self, this._then);

  final TrendTax _self;
  final $Res Function(TrendTax) _then;

/// Create a copy of TrendTax
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? ivaPayableRoundedCents = null,Object? isrPayableRoundedCents = null,Object? withholdingsToRemitCents = null,Object? totalPayableRoundedCents = null,Object? ivaOutcome = null,Object? status = null,}) {
  return _then(TrendTax(
ivaPayableRoundedCents: null == ivaPayableRoundedCents ? _self.ivaPayableRoundedCents : ivaPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,isrPayableRoundedCents: null == isrPayableRoundedCents ? _self.isrPayableRoundedCents : isrPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,withholdingsToRemitCents: null == withholdingsToRemitCents ? _self.withholdingsToRemitCents : withholdingsToRemitCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,totalPayableRoundedCents: null == totalPayableRoundedCents ? _self.totalPayableRoundedCents : totalPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,ivaOutcome: null == ivaOutcome ? _self.ivaOutcome : ivaOutcome // ignore: cast_nullable_to_non_nullable
as TrendTaxIvaOutcome,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as TaxpayerEstimateStatus,
  ));
}

}


/// Adds pattern-matching-related methods to [TrendTax].
extension TrendTaxPatterns on TrendTax {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _TrendTax value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _TrendTax() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _TrendTax value)  $default,){
final _that = this;
switch (_that) {
case _TrendTax():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _TrendTax value)?  $default,){
final _that = this;
switch (_that) {
case _TrendTax() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'iva_payable_rounded_cents')  NonNegativeCents ivaPayableRoundedCents, @JsonKey(name: 'isr_payable_rounded_cents')  NonNegativeCents isrPayableRoundedCents, @JsonKey(name: 'withholdings_to_remit_cents')  NonNegativeCents withholdingsToRemitCents, @JsonKey(name: 'total_payable_rounded_cents')  NonNegativeCents totalPayableRoundedCents, @JsonKey(name: 'iva_outcome')  TrendTaxIvaOutcome ivaOutcome,  TaxpayerEstimateStatus status)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TrendTax() when $default != null:
return $default(_that.ivaPayableRoundedCents,_that.isrPayableRoundedCents,_that.withholdingsToRemitCents,_that.totalPayableRoundedCents,_that.ivaOutcome,_that.status);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'iva_payable_rounded_cents')  NonNegativeCents ivaPayableRoundedCents, @JsonKey(name: 'isr_payable_rounded_cents')  NonNegativeCents isrPayableRoundedCents, @JsonKey(name: 'withholdings_to_remit_cents')  NonNegativeCents withholdingsToRemitCents, @JsonKey(name: 'total_payable_rounded_cents')  NonNegativeCents totalPayableRoundedCents, @JsonKey(name: 'iva_outcome')  TrendTaxIvaOutcome ivaOutcome,  TaxpayerEstimateStatus status)  $default,) {final _that = this;
switch (_that) {
case _TrendTax():
return $default(_that.ivaPayableRoundedCents,_that.isrPayableRoundedCents,_that.withholdingsToRemitCents,_that.totalPayableRoundedCents,_that.ivaOutcome,_that.status);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'iva_payable_rounded_cents')  NonNegativeCents ivaPayableRoundedCents, @JsonKey(name: 'isr_payable_rounded_cents')  NonNegativeCents isrPayableRoundedCents, @JsonKey(name: 'withholdings_to_remit_cents')  NonNegativeCents withholdingsToRemitCents, @JsonKey(name: 'total_payable_rounded_cents')  NonNegativeCents totalPayableRoundedCents, @JsonKey(name: 'iva_outcome')  TrendTaxIvaOutcome ivaOutcome,  TaxpayerEstimateStatus status)?  $default,) {final _that = this;
switch (_that) {
case _TrendTax() when $default != null:
return $default(_that.ivaPayableRoundedCents,_that.isrPayableRoundedCents,_that.withholdingsToRemitCents,_that.totalPayableRoundedCents,_that.ivaOutcome,_that.status);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TrendTax implements TrendTax {
  const _TrendTax({@JsonKey(name: 'iva_payable_rounded_cents') required this.ivaPayableRoundedCents, @JsonKey(name: 'isr_payable_rounded_cents') required this.isrPayableRoundedCents, @JsonKey(name: 'withholdings_to_remit_cents') required this.withholdingsToRemitCents, @JsonKey(name: 'total_payable_rounded_cents') required this.totalPayableRoundedCents, @JsonKey(name: 'iva_outcome') required this.ivaOutcome, required this.status});
  factory _TrendTax.fromJson(Map<String, dynamic> json) => _$TrendTaxFromJson(json);

@override@JsonKey(name: 'iva_payable_rounded_cents') final  NonNegativeCents ivaPayableRoundedCents;
@override@JsonKey(name: 'isr_payable_rounded_cents') final  NonNegativeCents isrPayableRoundedCents;
@override@JsonKey(name: 'withholdings_to_remit_cents') final  NonNegativeCents withholdingsToRemitCents;
@override@JsonKey(name: 'total_payable_rounded_cents') final  NonNegativeCents totalPayableRoundedCents;
@override@JsonKey(name: 'iva_outcome') final  TrendTaxIvaOutcome ivaOutcome;
@override final  TaxpayerEstimateStatus status;

/// Create a copy of TrendTax
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TrendTaxCopyWith<_TrendTax> get copyWith => __$TrendTaxCopyWithImpl<_TrendTax>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TrendTaxToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TrendTax&&(identical(other.ivaPayableRoundedCents, ivaPayableRoundedCents) || other.ivaPayableRoundedCents == ivaPayableRoundedCents)&&(identical(other.isrPayableRoundedCents, isrPayableRoundedCents) || other.isrPayableRoundedCents == isrPayableRoundedCents)&&(identical(other.withholdingsToRemitCents, withholdingsToRemitCents) || other.withholdingsToRemitCents == withholdingsToRemitCents)&&(identical(other.totalPayableRoundedCents, totalPayableRoundedCents) || other.totalPayableRoundedCents == totalPayableRoundedCents)&&(identical(other.ivaOutcome, ivaOutcome) || other.ivaOutcome == ivaOutcome)&&(identical(other.status, status) || other.status == status));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,ivaPayableRoundedCents,isrPayableRoundedCents,withholdingsToRemitCents,totalPayableRoundedCents,ivaOutcome,status);
}

@override
String toString() {
    return 'TrendTax(ivaPayableRoundedCents: $ivaPayableRoundedCents, isrPayableRoundedCents: $isrPayableRoundedCents, withholdingsToRemitCents: $withholdingsToRemitCents, totalPayableRoundedCents: $totalPayableRoundedCents, ivaOutcome: $ivaOutcome, status: $status)';
}


}

/// @nodoc
abstract mixin class _$TrendTaxCopyWith<$Res> implements $TrendTaxCopyWith<$Res> {
  factory _$TrendTaxCopyWith(_TrendTax value, $Res Function(_TrendTax) _then) = __$TrendTaxCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'iva_payable_rounded_cents') NonNegativeCents ivaPayableRoundedCents,@JsonKey(name: 'isr_payable_rounded_cents') NonNegativeCents isrPayableRoundedCents,@JsonKey(name: 'withholdings_to_remit_cents') NonNegativeCents withholdingsToRemitCents,@JsonKey(name: 'total_payable_rounded_cents') NonNegativeCents totalPayableRoundedCents,@JsonKey(name: 'iva_outcome') TrendTaxIvaOutcome ivaOutcome, TaxpayerEstimateStatus status
});




}
/// @nodoc
class __$TrendTaxCopyWithImpl<$Res>
    implements _$TrendTaxCopyWith<$Res> {
  __$TrendTaxCopyWithImpl(this._self, this._then);

  final _TrendTax _self;
  final $Res Function(_TrendTax) _then;

/// Create a copy of TrendTax
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? ivaPayableRoundedCents = null,Object? isrPayableRoundedCents = null,Object? withholdingsToRemitCents = null,Object? totalPayableRoundedCents = null,Object? ivaOutcome = null,Object? status = null,}) {
  return _then(_TrendTax(
ivaPayableRoundedCents: null == ivaPayableRoundedCents ? _self.ivaPayableRoundedCents : ivaPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,isrPayableRoundedCents: null == isrPayableRoundedCents ? _self.isrPayableRoundedCents : isrPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,withholdingsToRemitCents: null == withholdingsToRemitCents ? _self.withholdingsToRemitCents : withholdingsToRemitCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,totalPayableRoundedCents: null == totalPayableRoundedCents ? _self.totalPayableRoundedCents : totalPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,ivaOutcome: null == ivaOutcome ? _self.ivaOutcome : ivaOutcome // ignore: cast_nullable_to_non_nullable
as TrendTaxIvaOutcome,status: null == status ? _self.status : status // ignore: cast_nullable_to_non_nullable
as TaxpayerEstimateStatus,
  ));
}


}

// dart format on
