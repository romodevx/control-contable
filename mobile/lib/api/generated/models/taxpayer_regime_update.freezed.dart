// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'taxpayer_regime_update.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$TaxpayerRegimeUpdate {

 RegimeOptions get options;
/// Create a copy of TaxpayerRegimeUpdate
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TaxpayerRegimeUpdateCopyWith<TaxpayerRegimeUpdate> get copyWith => _$TaxpayerRegimeUpdateCopyWithImpl<TaxpayerRegimeUpdate>(this as TaxpayerRegimeUpdate, _$identity);

  /// Serializes this TaxpayerRegimeUpdate to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as TaxpayerRegimeUpdate;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TaxpayerRegimeUpdate&&const DeepCollectionEquality().equals(other.options, _this.options));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TaxpayerRegimeUpdate;
  return Object.hash(runtimeType,const DeepCollectionEquality().hash(_this.options));
}

@override
String toString() {
  final _this = this as TaxpayerRegimeUpdate;
  return 'TaxpayerRegimeUpdate(options: ${_this.options})';
}


}

/// @nodoc
abstract mixin class $TaxpayerRegimeUpdateCopyWith<$Res>  {
  factory $TaxpayerRegimeUpdateCopyWith(TaxpayerRegimeUpdate value, $Res Function(TaxpayerRegimeUpdate) _then) = _$TaxpayerRegimeUpdateCopyWithImpl;
@useResult
$Res call({
 RegimeOptions options
});




}
/// @nodoc
class _$TaxpayerRegimeUpdateCopyWithImpl<$Res>
    implements $TaxpayerRegimeUpdateCopyWith<$Res> {
  _$TaxpayerRegimeUpdateCopyWithImpl(this._self, this._then);

  final TaxpayerRegimeUpdate _self;
  final $Res Function(TaxpayerRegimeUpdate) _then;

/// Create a copy of TaxpayerRegimeUpdate
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? options = null,}) {
  return _then(TaxpayerRegimeUpdate(
options: null == options ? _self.options : options // ignore: cast_nullable_to_non_nullable
as RegimeOptions,
  ));
}

}


/// Adds pattern-matching-related methods to [TaxpayerRegimeUpdate].
extension TaxpayerRegimeUpdatePatterns on TaxpayerRegimeUpdate {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _TaxpayerRegimeUpdate value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _TaxpayerRegimeUpdate() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _TaxpayerRegimeUpdate value)  $default,){
final _that = this;
switch (_that) {
case _TaxpayerRegimeUpdate():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _TaxpayerRegimeUpdate value)?  $default,){
final _that = this;
switch (_that) {
case _TaxpayerRegimeUpdate() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( RegimeOptions options)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TaxpayerRegimeUpdate() when $default != null:
return $default(_that.options);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( RegimeOptions options)  $default,) {final _that = this;
switch (_that) {
case _TaxpayerRegimeUpdate():
return $default(_that.options);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( RegimeOptions options)?  $default,) {final _that = this;
switch (_that) {
case _TaxpayerRegimeUpdate() when $default != null:
return $default(_that.options);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TaxpayerRegimeUpdate implements TaxpayerRegimeUpdate {
  const _TaxpayerRegimeUpdate({required  RegimeOptions options}): _options = options;
  factory _TaxpayerRegimeUpdate.fromJson(Map<String, dynamic> json) => _$TaxpayerRegimeUpdateFromJson(json);

 final  RegimeOptions _options;
@override RegimeOptions get options {
  if (_options is EqualUnmodifiableMapView) return _options;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableMapView(_options);
}


/// Create a copy of TaxpayerRegimeUpdate
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TaxpayerRegimeUpdateCopyWith<_TaxpayerRegimeUpdate> get copyWith => __$TaxpayerRegimeUpdateCopyWithImpl<_TaxpayerRegimeUpdate>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TaxpayerRegimeUpdateToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TaxpayerRegimeUpdate&&const DeepCollectionEquality().equals(other.options, _options));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,const DeepCollectionEquality().hash(_options));
}

@override
String toString() {
    return 'TaxpayerRegimeUpdate(options: $options)';
}


}

/// @nodoc
abstract mixin class _$TaxpayerRegimeUpdateCopyWith<$Res> implements $TaxpayerRegimeUpdateCopyWith<$Res> {
  factory _$TaxpayerRegimeUpdateCopyWith(_TaxpayerRegimeUpdate value, $Res Function(_TaxpayerRegimeUpdate) _then) = __$TaxpayerRegimeUpdateCopyWithImpl;
@override @useResult
$Res call({
 RegimeOptions options
});




}
/// @nodoc
class __$TaxpayerRegimeUpdateCopyWithImpl<$Res>
    implements _$TaxpayerRegimeUpdateCopyWith<$Res> {
  __$TaxpayerRegimeUpdateCopyWithImpl(this._self, this._then);

  final _TaxpayerRegimeUpdate _self;
  final $Res Function(_TaxpayerRegimeUpdate) _then;

/// Create a copy of TaxpayerRegimeUpdate
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? options = null,}) {
  return _then(_TaxpayerRegimeUpdate(
options: null == options ? _self._options : options // ignore: cast_nullable_to_non_nullable
as RegimeOptions,
  ));
}


}

// dart format on
