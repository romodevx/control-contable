// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'history_tax.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$HistoryTax {

@JsonKey(name: 'isr_payable_rounded_cents') NonNegativeCents get isrPayableRoundedCents;@JsonKey(name: 'iva_payable_rounded_cents') NonNegativeCents get ivaPayableRoundedCents;@JsonKey(name: 'iva_outcome') HistoryTaxIvaOutcome get ivaOutcome;/// Saldo a favor of the consolidated IVA (the magnitude of a negative `totals.iva.amount_cents`); `0` unless `iva_outcome` is `in_favor`.
@JsonKey(name: 'iva_in_favor_cents') NonNegativeCents get ivaInFavorCents;@JsonKey(name: 'withholdings_to_remit_cents') NonNegativeCents get withholdingsToRemitCents;@JsonKey(name: 'total_payable_rounded_cents') NonNegativeCents get totalPayableRoundedCents;
/// Create a copy of HistoryTax
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$HistoryTaxCopyWith<HistoryTax> get copyWith => _$HistoryTaxCopyWithImpl<HistoryTax>(this as HistoryTax, _$identity);

  /// Serializes this HistoryTax to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as HistoryTax;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is HistoryTax&&(identical(other.isrPayableRoundedCents, _this.isrPayableRoundedCents) || other.isrPayableRoundedCents == _this.isrPayableRoundedCents)&&(identical(other.ivaPayableRoundedCents, _this.ivaPayableRoundedCents) || other.ivaPayableRoundedCents == _this.ivaPayableRoundedCents)&&(identical(other.ivaOutcome, _this.ivaOutcome) || other.ivaOutcome == _this.ivaOutcome)&&(identical(other.ivaInFavorCents, _this.ivaInFavorCents) || other.ivaInFavorCents == _this.ivaInFavorCents)&&(identical(other.withholdingsToRemitCents, _this.withholdingsToRemitCents) || other.withholdingsToRemitCents == _this.withholdingsToRemitCents)&&(identical(other.totalPayableRoundedCents, _this.totalPayableRoundedCents) || other.totalPayableRoundedCents == _this.totalPayableRoundedCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as HistoryTax;
  return Object.hash(runtimeType,_this.isrPayableRoundedCents,_this.ivaPayableRoundedCents,_this.ivaOutcome,_this.ivaInFavorCents,_this.withholdingsToRemitCents,_this.totalPayableRoundedCents);
}

@override
String toString() {
  final _this = this as HistoryTax;
  return 'HistoryTax(isrPayableRoundedCents: ${_this.isrPayableRoundedCents}, ivaPayableRoundedCents: ${_this.ivaPayableRoundedCents}, ivaOutcome: ${_this.ivaOutcome}, ivaInFavorCents: ${_this.ivaInFavorCents}, withholdingsToRemitCents: ${_this.withholdingsToRemitCents}, totalPayableRoundedCents: ${_this.totalPayableRoundedCents})';
}


}

/// @nodoc
abstract mixin class $HistoryTaxCopyWith<$Res>  {
  factory $HistoryTaxCopyWith(HistoryTax value, $Res Function(HistoryTax) _then) = _$HistoryTaxCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'isr_payable_rounded_cents') NonNegativeCents isrPayableRoundedCents,@JsonKey(name: 'iva_payable_rounded_cents') NonNegativeCents ivaPayableRoundedCents,@JsonKey(name: 'iva_outcome') HistoryTaxIvaOutcome ivaOutcome,@JsonKey(name: 'iva_in_favor_cents') NonNegativeCents ivaInFavorCents,@JsonKey(name: 'withholdings_to_remit_cents') NonNegativeCents withholdingsToRemitCents,@JsonKey(name: 'total_payable_rounded_cents') NonNegativeCents totalPayableRoundedCents
});




}
/// @nodoc
class _$HistoryTaxCopyWithImpl<$Res>
    implements $HistoryTaxCopyWith<$Res> {
  _$HistoryTaxCopyWithImpl(this._self, this._then);

  final HistoryTax _self;
  final $Res Function(HistoryTax) _then;

/// Create a copy of HistoryTax
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? isrPayableRoundedCents = null,Object? ivaPayableRoundedCents = null,Object? ivaOutcome = null,Object? ivaInFavorCents = null,Object? withholdingsToRemitCents = null,Object? totalPayableRoundedCents = null,}) {
  return _then(HistoryTax(
isrPayableRoundedCents: null == isrPayableRoundedCents ? _self.isrPayableRoundedCents : isrPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,ivaPayableRoundedCents: null == ivaPayableRoundedCents ? _self.ivaPayableRoundedCents : ivaPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,ivaOutcome: null == ivaOutcome ? _self.ivaOutcome : ivaOutcome // ignore: cast_nullable_to_non_nullable
as HistoryTaxIvaOutcome,ivaInFavorCents: null == ivaInFavorCents ? _self.ivaInFavorCents : ivaInFavorCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,withholdingsToRemitCents: null == withholdingsToRemitCents ? _self.withholdingsToRemitCents : withholdingsToRemitCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,totalPayableRoundedCents: null == totalPayableRoundedCents ? _self.totalPayableRoundedCents : totalPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,
  ));
}

}


/// Adds pattern-matching-related methods to [HistoryTax].
extension HistoryTaxPatterns on HistoryTax {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _HistoryTax value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _HistoryTax() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _HistoryTax value)  $default,){
final _that = this;
switch (_that) {
case _HistoryTax():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _HistoryTax value)?  $default,){
final _that = this;
switch (_that) {
case _HistoryTax() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'isr_payable_rounded_cents')  NonNegativeCents isrPayableRoundedCents, @JsonKey(name: 'iva_payable_rounded_cents')  NonNegativeCents ivaPayableRoundedCents, @JsonKey(name: 'iva_outcome')  HistoryTaxIvaOutcome ivaOutcome, @JsonKey(name: 'iva_in_favor_cents')  NonNegativeCents ivaInFavorCents, @JsonKey(name: 'withholdings_to_remit_cents')  NonNegativeCents withholdingsToRemitCents, @JsonKey(name: 'total_payable_rounded_cents')  NonNegativeCents totalPayableRoundedCents)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _HistoryTax() when $default != null:
return $default(_that.isrPayableRoundedCents,_that.ivaPayableRoundedCents,_that.ivaOutcome,_that.ivaInFavorCents,_that.withholdingsToRemitCents,_that.totalPayableRoundedCents);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'isr_payable_rounded_cents')  NonNegativeCents isrPayableRoundedCents, @JsonKey(name: 'iva_payable_rounded_cents')  NonNegativeCents ivaPayableRoundedCents, @JsonKey(name: 'iva_outcome')  HistoryTaxIvaOutcome ivaOutcome, @JsonKey(name: 'iva_in_favor_cents')  NonNegativeCents ivaInFavorCents, @JsonKey(name: 'withholdings_to_remit_cents')  NonNegativeCents withholdingsToRemitCents, @JsonKey(name: 'total_payable_rounded_cents')  NonNegativeCents totalPayableRoundedCents)  $default,) {final _that = this;
switch (_that) {
case _HistoryTax():
return $default(_that.isrPayableRoundedCents,_that.ivaPayableRoundedCents,_that.ivaOutcome,_that.ivaInFavorCents,_that.withholdingsToRemitCents,_that.totalPayableRoundedCents);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'isr_payable_rounded_cents')  NonNegativeCents isrPayableRoundedCents, @JsonKey(name: 'iva_payable_rounded_cents')  NonNegativeCents ivaPayableRoundedCents, @JsonKey(name: 'iva_outcome')  HistoryTaxIvaOutcome ivaOutcome, @JsonKey(name: 'iva_in_favor_cents')  NonNegativeCents ivaInFavorCents, @JsonKey(name: 'withholdings_to_remit_cents')  NonNegativeCents withholdingsToRemitCents, @JsonKey(name: 'total_payable_rounded_cents')  NonNegativeCents totalPayableRoundedCents)?  $default,) {final _that = this;
switch (_that) {
case _HistoryTax() when $default != null:
return $default(_that.isrPayableRoundedCents,_that.ivaPayableRoundedCents,_that.ivaOutcome,_that.ivaInFavorCents,_that.withholdingsToRemitCents,_that.totalPayableRoundedCents);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _HistoryTax implements HistoryTax {
  const _HistoryTax({@JsonKey(name: 'isr_payable_rounded_cents') required this.isrPayableRoundedCents, @JsonKey(name: 'iva_payable_rounded_cents') required this.ivaPayableRoundedCents, @JsonKey(name: 'iva_outcome') required this.ivaOutcome, @JsonKey(name: 'iva_in_favor_cents') required this.ivaInFavorCents, @JsonKey(name: 'withholdings_to_remit_cents') required this.withholdingsToRemitCents, @JsonKey(name: 'total_payable_rounded_cents') required this.totalPayableRoundedCents});
  factory _HistoryTax.fromJson(Map<String, dynamic> json) => _$HistoryTaxFromJson(json);

@override@JsonKey(name: 'isr_payable_rounded_cents') final  NonNegativeCents isrPayableRoundedCents;
@override@JsonKey(name: 'iva_payable_rounded_cents') final  NonNegativeCents ivaPayableRoundedCents;
@override@JsonKey(name: 'iva_outcome') final  HistoryTaxIvaOutcome ivaOutcome;
/// Saldo a favor of the consolidated IVA (the magnitude of a negative `totals.iva.amount_cents`); `0` unless `iva_outcome` is `in_favor`.
@override@JsonKey(name: 'iva_in_favor_cents') final  NonNegativeCents ivaInFavorCents;
@override@JsonKey(name: 'withholdings_to_remit_cents') final  NonNegativeCents withholdingsToRemitCents;
@override@JsonKey(name: 'total_payable_rounded_cents') final  NonNegativeCents totalPayableRoundedCents;

/// Create a copy of HistoryTax
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$HistoryTaxCopyWith<_HistoryTax> get copyWith => __$HistoryTaxCopyWithImpl<_HistoryTax>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$HistoryTaxToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _HistoryTax&&(identical(other.isrPayableRoundedCents, isrPayableRoundedCents) || other.isrPayableRoundedCents == isrPayableRoundedCents)&&(identical(other.ivaPayableRoundedCents, ivaPayableRoundedCents) || other.ivaPayableRoundedCents == ivaPayableRoundedCents)&&(identical(other.ivaOutcome, ivaOutcome) || other.ivaOutcome == ivaOutcome)&&(identical(other.ivaInFavorCents, ivaInFavorCents) || other.ivaInFavorCents == ivaInFavorCents)&&(identical(other.withholdingsToRemitCents, withholdingsToRemitCents) || other.withholdingsToRemitCents == withholdingsToRemitCents)&&(identical(other.totalPayableRoundedCents, totalPayableRoundedCents) || other.totalPayableRoundedCents == totalPayableRoundedCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,isrPayableRoundedCents,ivaPayableRoundedCents,ivaOutcome,ivaInFavorCents,withholdingsToRemitCents,totalPayableRoundedCents);
}

@override
String toString() {
    return 'HistoryTax(isrPayableRoundedCents: $isrPayableRoundedCents, ivaPayableRoundedCents: $ivaPayableRoundedCents, ivaOutcome: $ivaOutcome, ivaInFavorCents: $ivaInFavorCents, withholdingsToRemitCents: $withholdingsToRemitCents, totalPayableRoundedCents: $totalPayableRoundedCents)';
}


}

/// @nodoc
abstract mixin class _$HistoryTaxCopyWith<$Res> implements $HistoryTaxCopyWith<$Res> {
  factory _$HistoryTaxCopyWith(_HistoryTax value, $Res Function(_HistoryTax) _then) = __$HistoryTaxCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'isr_payable_rounded_cents') NonNegativeCents isrPayableRoundedCents,@JsonKey(name: 'iva_payable_rounded_cents') NonNegativeCents ivaPayableRoundedCents,@JsonKey(name: 'iva_outcome') HistoryTaxIvaOutcome ivaOutcome,@JsonKey(name: 'iva_in_favor_cents') NonNegativeCents ivaInFavorCents,@JsonKey(name: 'withholdings_to_remit_cents') NonNegativeCents withholdingsToRemitCents,@JsonKey(name: 'total_payable_rounded_cents') NonNegativeCents totalPayableRoundedCents
});




}
/// @nodoc
class __$HistoryTaxCopyWithImpl<$Res>
    implements _$HistoryTaxCopyWith<$Res> {
  __$HistoryTaxCopyWithImpl(this._self, this._then);

  final _HistoryTax _self;
  final $Res Function(_HistoryTax) _then;

/// Create a copy of HistoryTax
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? isrPayableRoundedCents = null,Object? ivaPayableRoundedCents = null,Object? ivaOutcome = null,Object? ivaInFavorCents = null,Object? withholdingsToRemitCents = null,Object? totalPayableRoundedCents = null,}) {
  return _then(_HistoryTax(
isrPayableRoundedCents: null == isrPayableRoundedCents ? _self.isrPayableRoundedCents : isrPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,ivaPayableRoundedCents: null == ivaPayableRoundedCents ? _self.ivaPayableRoundedCents : ivaPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,ivaOutcome: null == ivaOutcome ? _self.ivaOutcome : ivaOutcome // ignore: cast_nullable_to_non_nullable
as HistoryTaxIvaOutcome,ivaInFavorCents: null == ivaInFavorCents ? _self.ivaInFavorCents : ivaInFavorCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,withholdingsToRemitCents: null == withholdingsToRemitCents ? _self.withholdingsToRemitCents : withholdingsToRemitCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,totalPayableRoundedCents: null == totalPayableRoundedCents ? _self.totalPayableRoundedCents : totalPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,
  ));
}


}

// dart format on
