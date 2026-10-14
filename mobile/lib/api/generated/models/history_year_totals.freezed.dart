// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'history_year_totals.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$HistoryYearTotals {

@JsonKey(name: 'isr_payable_rounded_cents') NonNegativeCents get isrPayableRoundedCents;@JsonKey(name: 'iva_payable_rounded_cents') NonNegativeCents get ivaPayableRoundedCents;@JsonKey(name: 'withholdings_to_remit_cents') NonNegativeCents get withholdingsToRemitCents;@JsonKey(name: 'total_payable_rounded_cents') NonNegativeCents get totalPayableRoundedCents;
/// Create a copy of HistoryYearTotals
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$HistoryYearTotalsCopyWith<HistoryYearTotals> get copyWith => _$HistoryYearTotalsCopyWithImpl<HistoryYearTotals>(this as HistoryYearTotals, _$identity);

  /// Serializes this HistoryYearTotals to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as HistoryYearTotals;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is HistoryYearTotals&&(identical(other.isrPayableRoundedCents, _this.isrPayableRoundedCents) || other.isrPayableRoundedCents == _this.isrPayableRoundedCents)&&(identical(other.ivaPayableRoundedCents, _this.ivaPayableRoundedCents) || other.ivaPayableRoundedCents == _this.ivaPayableRoundedCents)&&(identical(other.withholdingsToRemitCents, _this.withholdingsToRemitCents) || other.withholdingsToRemitCents == _this.withholdingsToRemitCents)&&(identical(other.totalPayableRoundedCents, _this.totalPayableRoundedCents) || other.totalPayableRoundedCents == _this.totalPayableRoundedCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as HistoryYearTotals;
  return Object.hash(runtimeType,_this.isrPayableRoundedCents,_this.ivaPayableRoundedCents,_this.withholdingsToRemitCents,_this.totalPayableRoundedCents);
}

@override
String toString() {
  final _this = this as HistoryYearTotals;
  return 'HistoryYearTotals(isrPayableRoundedCents: ${_this.isrPayableRoundedCents}, ivaPayableRoundedCents: ${_this.ivaPayableRoundedCents}, withholdingsToRemitCents: ${_this.withholdingsToRemitCents}, totalPayableRoundedCents: ${_this.totalPayableRoundedCents})';
}


}

/// @nodoc
abstract mixin class $HistoryYearTotalsCopyWith<$Res>  {
  factory $HistoryYearTotalsCopyWith(HistoryYearTotals value, $Res Function(HistoryYearTotals) _then) = _$HistoryYearTotalsCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'isr_payable_rounded_cents') NonNegativeCents isrPayableRoundedCents,@JsonKey(name: 'iva_payable_rounded_cents') NonNegativeCents ivaPayableRoundedCents,@JsonKey(name: 'withholdings_to_remit_cents') NonNegativeCents withholdingsToRemitCents,@JsonKey(name: 'total_payable_rounded_cents') NonNegativeCents totalPayableRoundedCents
});




}
/// @nodoc
class _$HistoryYearTotalsCopyWithImpl<$Res>
    implements $HistoryYearTotalsCopyWith<$Res> {
  _$HistoryYearTotalsCopyWithImpl(this._self, this._then);

  final HistoryYearTotals _self;
  final $Res Function(HistoryYearTotals) _then;

/// Create a copy of HistoryYearTotals
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? isrPayableRoundedCents = null,Object? ivaPayableRoundedCents = null,Object? withholdingsToRemitCents = null,Object? totalPayableRoundedCents = null,}) {
  return _then(HistoryYearTotals(
isrPayableRoundedCents: null == isrPayableRoundedCents ? _self.isrPayableRoundedCents : isrPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,ivaPayableRoundedCents: null == ivaPayableRoundedCents ? _self.ivaPayableRoundedCents : ivaPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,withholdingsToRemitCents: null == withholdingsToRemitCents ? _self.withholdingsToRemitCents : withholdingsToRemitCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,totalPayableRoundedCents: null == totalPayableRoundedCents ? _self.totalPayableRoundedCents : totalPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,
  ));
}

}


/// Adds pattern-matching-related methods to [HistoryYearTotals].
extension HistoryYearTotalsPatterns on HistoryYearTotals {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _HistoryYearTotals value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _HistoryYearTotals() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _HistoryYearTotals value)  $default,){
final _that = this;
switch (_that) {
case _HistoryYearTotals():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _HistoryYearTotals value)?  $default,){
final _that = this;
switch (_that) {
case _HistoryYearTotals() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'isr_payable_rounded_cents')  NonNegativeCents isrPayableRoundedCents, @JsonKey(name: 'iva_payable_rounded_cents')  NonNegativeCents ivaPayableRoundedCents, @JsonKey(name: 'withholdings_to_remit_cents')  NonNegativeCents withholdingsToRemitCents, @JsonKey(name: 'total_payable_rounded_cents')  NonNegativeCents totalPayableRoundedCents)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _HistoryYearTotals() when $default != null:
return $default(_that.isrPayableRoundedCents,_that.ivaPayableRoundedCents,_that.withholdingsToRemitCents,_that.totalPayableRoundedCents);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'isr_payable_rounded_cents')  NonNegativeCents isrPayableRoundedCents, @JsonKey(name: 'iva_payable_rounded_cents')  NonNegativeCents ivaPayableRoundedCents, @JsonKey(name: 'withholdings_to_remit_cents')  NonNegativeCents withholdingsToRemitCents, @JsonKey(name: 'total_payable_rounded_cents')  NonNegativeCents totalPayableRoundedCents)  $default,) {final _that = this;
switch (_that) {
case _HistoryYearTotals():
return $default(_that.isrPayableRoundedCents,_that.ivaPayableRoundedCents,_that.withholdingsToRemitCents,_that.totalPayableRoundedCents);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'isr_payable_rounded_cents')  NonNegativeCents isrPayableRoundedCents, @JsonKey(name: 'iva_payable_rounded_cents')  NonNegativeCents ivaPayableRoundedCents, @JsonKey(name: 'withholdings_to_remit_cents')  NonNegativeCents withholdingsToRemitCents, @JsonKey(name: 'total_payable_rounded_cents')  NonNegativeCents totalPayableRoundedCents)?  $default,) {final _that = this;
switch (_that) {
case _HistoryYearTotals() when $default != null:
return $default(_that.isrPayableRoundedCents,_that.ivaPayableRoundedCents,_that.withholdingsToRemitCents,_that.totalPayableRoundedCents);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _HistoryYearTotals implements HistoryYearTotals {
  const _HistoryYearTotals({@JsonKey(name: 'isr_payable_rounded_cents') required this.isrPayableRoundedCents, @JsonKey(name: 'iva_payable_rounded_cents') required this.ivaPayableRoundedCents, @JsonKey(name: 'withholdings_to_remit_cents') required this.withholdingsToRemitCents, @JsonKey(name: 'total_payable_rounded_cents') required this.totalPayableRoundedCents});
  factory _HistoryYearTotals.fromJson(Map<String, dynamic> json) => _$HistoryYearTotalsFromJson(json);

@override@JsonKey(name: 'isr_payable_rounded_cents') final  NonNegativeCents isrPayableRoundedCents;
@override@JsonKey(name: 'iva_payable_rounded_cents') final  NonNegativeCents ivaPayableRoundedCents;
@override@JsonKey(name: 'withholdings_to_remit_cents') final  NonNegativeCents withholdingsToRemitCents;
@override@JsonKey(name: 'total_payable_rounded_cents') final  NonNegativeCents totalPayableRoundedCents;

/// Create a copy of HistoryYearTotals
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$HistoryYearTotalsCopyWith<_HistoryYearTotals> get copyWith => __$HistoryYearTotalsCopyWithImpl<_HistoryYearTotals>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$HistoryYearTotalsToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _HistoryYearTotals&&(identical(other.isrPayableRoundedCents, isrPayableRoundedCents) || other.isrPayableRoundedCents == isrPayableRoundedCents)&&(identical(other.ivaPayableRoundedCents, ivaPayableRoundedCents) || other.ivaPayableRoundedCents == ivaPayableRoundedCents)&&(identical(other.withholdingsToRemitCents, withholdingsToRemitCents) || other.withholdingsToRemitCents == withholdingsToRemitCents)&&(identical(other.totalPayableRoundedCents, totalPayableRoundedCents) || other.totalPayableRoundedCents == totalPayableRoundedCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,isrPayableRoundedCents,ivaPayableRoundedCents,withholdingsToRemitCents,totalPayableRoundedCents);
}

@override
String toString() {
    return 'HistoryYearTotals(isrPayableRoundedCents: $isrPayableRoundedCents, ivaPayableRoundedCents: $ivaPayableRoundedCents, withholdingsToRemitCents: $withholdingsToRemitCents, totalPayableRoundedCents: $totalPayableRoundedCents)';
}


}

/// @nodoc
abstract mixin class _$HistoryYearTotalsCopyWith<$Res> implements $HistoryYearTotalsCopyWith<$Res> {
  factory _$HistoryYearTotalsCopyWith(_HistoryYearTotals value, $Res Function(_HistoryYearTotals) _then) = __$HistoryYearTotalsCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'isr_payable_rounded_cents') NonNegativeCents isrPayableRoundedCents,@JsonKey(name: 'iva_payable_rounded_cents') NonNegativeCents ivaPayableRoundedCents,@JsonKey(name: 'withholdings_to_remit_cents') NonNegativeCents withholdingsToRemitCents,@JsonKey(name: 'total_payable_rounded_cents') NonNegativeCents totalPayableRoundedCents
});




}
/// @nodoc
class __$HistoryYearTotalsCopyWithImpl<$Res>
    implements _$HistoryYearTotalsCopyWith<$Res> {
  __$HistoryYearTotalsCopyWithImpl(this._self, this._then);

  final _HistoryYearTotals _self;
  final $Res Function(_HistoryYearTotals) _then;

/// Create a copy of HistoryYearTotals
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? isrPayableRoundedCents = null,Object? ivaPayableRoundedCents = null,Object? withholdingsToRemitCents = null,Object? totalPayableRoundedCents = null,}) {
  return _then(_HistoryYearTotals(
isrPayableRoundedCents: null == isrPayableRoundedCents ? _self.isrPayableRoundedCents : isrPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,ivaPayableRoundedCents: null == ivaPayableRoundedCents ? _self.ivaPayableRoundedCents : ivaPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,withholdingsToRemitCents: null == withholdingsToRemitCents ? _self.withholdingsToRemitCents : withholdingsToRemitCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,totalPayableRoundedCents: null == totalPayableRoundedCents ? _self.totalPayableRoundedCents : totalPayableRoundedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,
  ));
}


}

// dart format on
