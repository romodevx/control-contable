// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'iva_components.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$IvaComponents {

@JsonKey(name: 'charged_collected_cents') NonNegativeCents get chargedCollectedCents;@JsonKey(name: 'creditable_paid_cents') NonNegativeCents get creditablePaidCents;@JsonKey(name: 'withheld_from_taxpayer_cents') NonNegativeCents get withheldFromTaxpayerCents;/// IVA already definitive via withholding (e.g. 625 under LIVA 18-M).
@JsonKey(includeIfNull: false, name: 'definitive_excluded_cents') NonNegativeCents? get definitiveExcludedCents;
/// Create a copy of IvaComponents
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$IvaComponentsCopyWith<IvaComponents> get copyWith => _$IvaComponentsCopyWithImpl<IvaComponents>(this as IvaComponents, _$identity);

  /// Serializes this IvaComponents to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as IvaComponents;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is IvaComponents&&(identical(other.chargedCollectedCents, _this.chargedCollectedCents) || other.chargedCollectedCents == _this.chargedCollectedCents)&&(identical(other.creditablePaidCents, _this.creditablePaidCents) || other.creditablePaidCents == _this.creditablePaidCents)&&(identical(other.withheldFromTaxpayerCents, _this.withheldFromTaxpayerCents) || other.withheldFromTaxpayerCents == _this.withheldFromTaxpayerCents)&&(identical(other.definitiveExcludedCents, _this.definitiveExcludedCents) || other.definitiveExcludedCents == _this.definitiveExcludedCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as IvaComponents;
  return Object.hash(runtimeType,_this.chargedCollectedCents,_this.creditablePaidCents,_this.withheldFromTaxpayerCents,_this.definitiveExcludedCents);
}

@override
String toString() {
  final _this = this as IvaComponents;
  return 'IvaComponents(chargedCollectedCents: ${_this.chargedCollectedCents}, creditablePaidCents: ${_this.creditablePaidCents}, withheldFromTaxpayerCents: ${_this.withheldFromTaxpayerCents}, definitiveExcludedCents: ${_this.definitiveExcludedCents})';
}


}

/// @nodoc
abstract mixin class $IvaComponentsCopyWith<$Res>  {
  factory $IvaComponentsCopyWith(IvaComponents value, $Res Function(IvaComponents) _then) = _$IvaComponentsCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'charged_collected_cents') NonNegativeCents chargedCollectedCents,@JsonKey(name: 'creditable_paid_cents') NonNegativeCents creditablePaidCents,@JsonKey(name: 'withheld_from_taxpayer_cents') NonNegativeCents withheldFromTaxpayerCents,@JsonKey(includeIfNull: false, name: 'definitive_excluded_cents') NonNegativeCents? definitiveExcludedCents
});




}
/// @nodoc
class _$IvaComponentsCopyWithImpl<$Res>
    implements $IvaComponentsCopyWith<$Res> {
  _$IvaComponentsCopyWithImpl(this._self, this._then);

  final IvaComponents _self;
  final $Res Function(IvaComponents) _then;

/// Create a copy of IvaComponents
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? chargedCollectedCents = null,Object? creditablePaidCents = null,Object? withheldFromTaxpayerCents = null,Object? definitiveExcludedCents = freezed,}) {
  return _then(IvaComponents(
chargedCollectedCents: null == chargedCollectedCents ? _self.chargedCollectedCents : chargedCollectedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,creditablePaidCents: null == creditablePaidCents ? _self.creditablePaidCents : creditablePaidCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,withheldFromTaxpayerCents: null == withheldFromTaxpayerCents ? _self.withheldFromTaxpayerCents : withheldFromTaxpayerCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,definitiveExcludedCents: freezed == definitiveExcludedCents ? _self.definitiveExcludedCents : definitiveExcludedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,
  ));
}

}


/// Adds pattern-matching-related methods to [IvaComponents].
extension IvaComponentsPatterns on IvaComponents {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _IvaComponents value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _IvaComponents() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _IvaComponents value)  $default,){
final _that = this;
switch (_that) {
case _IvaComponents():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _IvaComponents value)?  $default,){
final _that = this;
switch (_that) {
case _IvaComponents() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'charged_collected_cents')  NonNegativeCents chargedCollectedCents, @JsonKey(name: 'creditable_paid_cents')  NonNegativeCents creditablePaidCents, @JsonKey(name: 'withheld_from_taxpayer_cents')  NonNegativeCents withheldFromTaxpayerCents, @JsonKey(includeIfNull: false, name: 'definitive_excluded_cents')  NonNegativeCents? definitiveExcludedCents)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _IvaComponents() when $default != null:
return $default(_that.chargedCollectedCents,_that.creditablePaidCents,_that.withheldFromTaxpayerCents,_that.definitiveExcludedCents);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'charged_collected_cents')  NonNegativeCents chargedCollectedCents, @JsonKey(name: 'creditable_paid_cents')  NonNegativeCents creditablePaidCents, @JsonKey(name: 'withheld_from_taxpayer_cents')  NonNegativeCents withheldFromTaxpayerCents, @JsonKey(includeIfNull: false, name: 'definitive_excluded_cents')  NonNegativeCents? definitiveExcludedCents)  $default,) {final _that = this;
switch (_that) {
case _IvaComponents():
return $default(_that.chargedCollectedCents,_that.creditablePaidCents,_that.withheldFromTaxpayerCents,_that.definitiveExcludedCents);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'charged_collected_cents')  NonNegativeCents chargedCollectedCents, @JsonKey(name: 'creditable_paid_cents')  NonNegativeCents creditablePaidCents, @JsonKey(name: 'withheld_from_taxpayer_cents')  NonNegativeCents withheldFromTaxpayerCents, @JsonKey(includeIfNull: false, name: 'definitive_excluded_cents')  NonNegativeCents? definitiveExcludedCents)?  $default,) {final _that = this;
switch (_that) {
case _IvaComponents() when $default != null:
return $default(_that.chargedCollectedCents,_that.creditablePaidCents,_that.withheldFromTaxpayerCents,_that.definitiveExcludedCents);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _IvaComponents implements IvaComponents {
  const _IvaComponents({@JsonKey(name: 'charged_collected_cents') required this.chargedCollectedCents, @JsonKey(name: 'creditable_paid_cents') required this.creditablePaidCents, @JsonKey(name: 'withheld_from_taxpayer_cents') required this.withheldFromTaxpayerCents, @JsonKey(includeIfNull: false, name: 'definitive_excluded_cents') this.definitiveExcludedCents});
  factory _IvaComponents.fromJson(Map<String, dynamic> json) => _$IvaComponentsFromJson(json);

@override@JsonKey(name: 'charged_collected_cents') final  NonNegativeCents chargedCollectedCents;
@override@JsonKey(name: 'creditable_paid_cents') final  NonNegativeCents creditablePaidCents;
@override@JsonKey(name: 'withheld_from_taxpayer_cents') final  NonNegativeCents withheldFromTaxpayerCents;
/// IVA already definitive via withholding (e.g. 625 under LIVA 18-M).
@override@JsonKey(includeIfNull: false, name: 'definitive_excluded_cents') final  NonNegativeCents? definitiveExcludedCents;

/// Create a copy of IvaComponents
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$IvaComponentsCopyWith<_IvaComponents> get copyWith => __$IvaComponentsCopyWithImpl<_IvaComponents>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$IvaComponentsToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _IvaComponents&&(identical(other.chargedCollectedCents, chargedCollectedCents) || other.chargedCollectedCents == chargedCollectedCents)&&(identical(other.creditablePaidCents, creditablePaidCents) || other.creditablePaidCents == creditablePaidCents)&&(identical(other.withheldFromTaxpayerCents, withheldFromTaxpayerCents) || other.withheldFromTaxpayerCents == withheldFromTaxpayerCents)&&(identical(other.definitiveExcludedCents, definitiveExcludedCents) || other.definitiveExcludedCents == definitiveExcludedCents));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,chargedCollectedCents,creditablePaidCents,withheldFromTaxpayerCents,definitiveExcludedCents);
}

@override
String toString() {
    return 'IvaComponents(chargedCollectedCents: $chargedCollectedCents, creditablePaidCents: $creditablePaidCents, withheldFromTaxpayerCents: $withheldFromTaxpayerCents, definitiveExcludedCents: $definitiveExcludedCents)';
}


}

/// @nodoc
abstract mixin class _$IvaComponentsCopyWith<$Res> implements $IvaComponentsCopyWith<$Res> {
  factory _$IvaComponentsCopyWith(_IvaComponents value, $Res Function(_IvaComponents) _then) = __$IvaComponentsCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'charged_collected_cents') NonNegativeCents chargedCollectedCents,@JsonKey(name: 'creditable_paid_cents') NonNegativeCents creditablePaidCents,@JsonKey(name: 'withheld_from_taxpayer_cents') NonNegativeCents withheldFromTaxpayerCents,@JsonKey(includeIfNull: false, name: 'definitive_excluded_cents') NonNegativeCents? definitiveExcludedCents
});




}
/// @nodoc
class __$IvaComponentsCopyWithImpl<$Res>
    implements _$IvaComponentsCopyWith<$Res> {
  __$IvaComponentsCopyWithImpl(this._self, this._then);

  final _IvaComponents _self;
  final $Res Function(_IvaComponents) _then;

/// Create a copy of IvaComponents
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? chargedCollectedCents = null,Object? creditablePaidCents = null,Object? withheldFromTaxpayerCents = null,Object? definitiveExcludedCents = freezed,}) {
  return _then(_IvaComponents(
chargedCollectedCents: null == chargedCollectedCents ? _self.chargedCollectedCents : chargedCollectedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,creditablePaidCents: null == creditablePaidCents ? _self.creditablePaidCents : creditablePaidCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,withheldFromTaxpayerCents: null == withheldFromTaxpayerCents ? _self.withheldFromTaxpayerCents : withheldFromTaxpayerCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,definitiveExcludedCents: freezed == definitiveExcludedCents ? _self.definitiveExcludedCents : definitiveExcludedCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,
  ));
}


}

// dart format on
