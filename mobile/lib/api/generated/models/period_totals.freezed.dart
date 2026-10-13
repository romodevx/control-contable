// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'period_totals.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$PeriodTotals {

 int get count;@JsonKey(name: 'amount_cents') NonNegativeCents get amountCents;@JsonKey(name: 'iva_cents') NonNegativeCents get ivaCents;@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') NonNegativeCents? get isrWithheldCents;@JsonKey(includeIfNull: false, name: 'iva_withheld_cents') NonNegativeCents? get ivaWithheldCents;
/// Create a copy of PeriodTotals
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$PeriodTotalsCopyWith<PeriodTotals> get copyWith => _$PeriodTotalsCopyWithImpl<PeriodTotals>(this as PeriodTotals, _$identity);

  /// Serializes this PeriodTotals to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as PeriodTotals;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is PeriodTotals&&(identical(other.count, _this.count) || other.count == _this.count)&&(identical(other.amountCents, _this.amountCents) || other.amountCents == _this.amountCents)&&(identical(other.ivaCents, _this.ivaCents) || other.ivaCents == _this.ivaCents)&&(identical(other.isrWithheldCents, _this.isrWithheldCents) || other.isrWithheldCents == _this.isrWithheldCents)&&(identical(other.ivaWithheldCents, _this.ivaWithheldCents) || other.ivaWithheldCents == _this.ivaWithheldCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as PeriodTotals;
  return Object.hash(runtimeType,_this.count,_this.amountCents,_this.ivaCents,_this.isrWithheldCents,_this.ivaWithheldCents);
}

@override
String toString() {
  final _this = this as PeriodTotals;
  return 'PeriodTotals(count: ${_this.count}, amountCents: ${_this.amountCents}, ivaCents: ${_this.ivaCents}, isrWithheldCents: ${_this.isrWithheldCents}, ivaWithheldCents: ${_this.ivaWithheldCents})';
}


}

/// @nodoc
abstract mixin class $PeriodTotalsCopyWith<$Res>  {
  factory $PeriodTotalsCopyWith(PeriodTotals value, $Res Function(PeriodTotals) _then) = _$PeriodTotalsCopyWithImpl;
@useResult
$Res call({
 int count,@JsonKey(name: 'amount_cents') NonNegativeCents amountCents,@JsonKey(name: 'iva_cents') NonNegativeCents ivaCents,@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') NonNegativeCents? isrWithheldCents,@JsonKey(includeIfNull: false, name: 'iva_withheld_cents') NonNegativeCents? ivaWithheldCents
});




}
/// @nodoc
class _$PeriodTotalsCopyWithImpl<$Res>
    implements $PeriodTotalsCopyWith<$Res> {
  _$PeriodTotalsCopyWithImpl(this._self, this._then);

  final PeriodTotals _self;
  final $Res Function(PeriodTotals) _then;

/// Create a copy of PeriodTotals
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? count = null,Object? amountCents = null,Object? ivaCents = null,Object? isrWithheldCents = freezed,Object? ivaWithheldCents = freezed,}) {
  return _then(PeriodTotals(
count: null == count ? _self.count : count // ignore: cast_nullable_to_non_nullable
as int,amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,ivaCents: null == ivaCents ? _self.ivaCents : ivaCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,isrWithheldCents: freezed == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,ivaWithheldCents: freezed == ivaWithheldCents ? _self.ivaWithheldCents : ivaWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,
  ));
}

}


/// Adds pattern-matching-related methods to [PeriodTotals].
extension PeriodTotalsPatterns on PeriodTotals {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _PeriodTotals value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _PeriodTotals() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _PeriodTotals value)  $default,){
final _that = this;
switch (_that) {
case _PeriodTotals():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _PeriodTotals value)?  $default,){
final _that = this;
switch (_that) {
case _PeriodTotals() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( int count, @JsonKey(name: 'amount_cents')  NonNegativeCents amountCents, @JsonKey(name: 'iva_cents')  NonNegativeCents ivaCents, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(includeIfNull: false, name: 'iva_withheld_cents')  NonNegativeCents? ivaWithheldCents)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _PeriodTotals() when $default != null:
return $default(_that.count,_that.amountCents,_that.ivaCents,_that.isrWithheldCents,_that.ivaWithheldCents);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( int count, @JsonKey(name: 'amount_cents')  NonNegativeCents amountCents, @JsonKey(name: 'iva_cents')  NonNegativeCents ivaCents, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(includeIfNull: false, name: 'iva_withheld_cents')  NonNegativeCents? ivaWithheldCents)  $default,) {final _that = this;
switch (_that) {
case _PeriodTotals():
return $default(_that.count,_that.amountCents,_that.ivaCents,_that.isrWithheldCents,_that.ivaWithheldCents);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( int count, @JsonKey(name: 'amount_cents')  NonNegativeCents amountCents, @JsonKey(name: 'iva_cents')  NonNegativeCents ivaCents, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents')  NonNegativeCents? isrWithheldCents, @JsonKey(includeIfNull: false, name: 'iva_withheld_cents')  NonNegativeCents? ivaWithheldCents)?  $default,) {final _that = this;
switch (_that) {
case _PeriodTotals() when $default != null:
return $default(_that.count,_that.amountCents,_that.ivaCents,_that.isrWithheldCents,_that.ivaWithheldCents);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _PeriodTotals implements PeriodTotals {
  const _PeriodTotals({required this.count, @JsonKey(name: 'amount_cents') required this.amountCents, @JsonKey(name: 'iva_cents') required this.ivaCents, @JsonKey(includeIfNull: false, name: 'isr_withheld_cents') this.isrWithheldCents, @JsonKey(includeIfNull: false, name: 'iva_withheld_cents') this.ivaWithheldCents});
  factory _PeriodTotals.fromJson(Map<String, dynamic> json) => _$PeriodTotalsFromJson(json);

@override final  int count;
@override@JsonKey(name: 'amount_cents') final  NonNegativeCents amountCents;
@override@JsonKey(name: 'iva_cents') final  NonNegativeCents ivaCents;
@override@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') final  NonNegativeCents? isrWithheldCents;
@override@JsonKey(includeIfNull: false, name: 'iva_withheld_cents') final  NonNegativeCents? ivaWithheldCents;

/// Create a copy of PeriodTotals
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$PeriodTotalsCopyWith<_PeriodTotals> get copyWith => __$PeriodTotalsCopyWithImpl<_PeriodTotals>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$PeriodTotalsToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _PeriodTotals&&(identical(other.count, count) || other.count == count)&&(identical(other.amountCents, amountCents) || other.amountCents == amountCents)&&(identical(other.ivaCents, ivaCents) || other.ivaCents == ivaCents)&&(identical(other.isrWithheldCents, isrWithheldCents) || other.isrWithheldCents == isrWithheldCents)&&(identical(other.ivaWithheldCents, ivaWithheldCents) || other.ivaWithheldCents == ivaWithheldCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,count,amountCents,ivaCents,isrWithheldCents,ivaWithheldCents);
}

@override
String toString() {
    return 'PeriodTotals(count: $count, amountCents: $amountCents, ivaCents: $ivaCents, isrWithheldCents: $isrWithheldCents, ivaWithheldCents: $ivaWithheldCents)';
}


}

/// @nodoc
abstract mixin class _$PeriodTotalsCopyWith<$Res> implements $PeriodTotalsCopyWith<$Res> {
  factory _$PeriodTotalsCopyWith(_PeriodTotals value, $Res Function(_PeriodTotals) _then) = __$PeriodTotalsCopyWithImpl;
@override @useResult
$Res call({
 int count,@JsonKey(name: 'amount_cents') NonNegativeCents amountCents,@JsonKey(name: 'iva_cents') NonNegativeCents ivaCents,@JsonKey(includeIfNull: false, name: 'isr_withheld_cents') NonNegativeCents? isrWithheldCents,@JsonKey(includeIfNull: false, name: 'iva_withheld_cents') NonNegativeCents? ivaWithheldCents
});




}
/// @nodoc
class __$PeriodTotalsCopyWithImpl<$Res>
    implements _$PeriodTotalsCopyWith<$Res> {
  __$PeriodTotalsCopyWithImpl(this._self, this._then);

  final _PeriodTotals _self;
  final $Res Function(_PeriodTotals) _then;

/// Create a copy of PeriodTotals
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? count = null,Object? amountCents = null,Object? ivaCents = null,Object? isrWithheldCents = freezed,Object? ivaWithheldCents = freezed,}) {
  return _then(_PeriodTotals(
count: null == count ? _self.count : count // ignore: cast_nullable_to_non_nullable
as int,amountCents: null == amountCents ? _self.amountCents : amountCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,ivaCents: null == ivaCents ? _self.ivaCents : ivaCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,isrWithheldCents: freezed == isrWithheldCents ? _self.isrWithheldCents : isrWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,ivaWithheldCents: freezed == ivaWithheldCents ? _self.ivaWithheldCents : ivaWithheldCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,
  ));
}


}

// dart format on
