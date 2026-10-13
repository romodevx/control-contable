// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'taxpayer_update.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$TaxpayerUpdate {

@JsonKey(name: 'display_name') String get displayName;
/// Create a copy of TaxpayerUpdate
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TaxpayerUpdateCopyWith<TaxpayerUpdate> get copyWith => _$TaxpayerUpdateCopyWithImpl<TaxpayerUpdate>(this as TaxpayerUpdate, _$identity);

  /// Serializes this TaxpayerUpdate to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as TaxpayerUpdate;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TaxpayerUpdate&&(identical(other.displayName, _this.displayName) || other.displayName == _this.displayName));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TaxpayerUpdate;
  return Object.hash(runtimeType,_this.displayName);
}

@override
String toString() {
  final _this = this as TaxpayerUpdate;
  return 'TaxpayerUpdate(displayName: ${_this.displayName})';
}


}

/// @nodoc
abstract mixin class $TaxpayerUpdateCopyWith<$Res>  {
  factory $TaxpayerUpdateCopyWith(TaxpayerUpdate value, $Res Function(TaxpayerUpdate) _then) = _$TaxpayerUpdateCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'display_name') String displayName
});




}
/// @nodoc
class _$TaxpayerUpdateCopyWithImpl<$Res>
    implements $TaxpayerUpdateCopyWith<$Res> {
  _$TaxpayerUpdateCopyWithImpl(this._self, this._then);

  final TaxpayerUpdate _self;
  final $Res Function(TaxpayerUpdate) _then;

/// Create a copy of TaxpayerUpdate
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? displayName = null,}) {
  return _then(TaxpayerUpdate(
displayName: null == displayName ? _self.displayName : displayName // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [TaxpayerUpdate].
extension TaxpayerUpdatePatterns on TaxpayerUpdate {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _TaxpayerUpdate value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _TaxpayerUpdate() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _TaxpayerUpdate value)  $default,){
final _that = this;
switch (_that) {
case _TaxpayerUpdate():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _TaxpayerUpdate value)?  $default,){
final _that = this;
switch (_that) {
case _TaxpayerUpdate() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'display_name')  String displayName)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TaxpayerUpdate() when $default != null:
return $default(_that.displayName);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'display_name')  String displayName)  $default,) {final _that = this;
switch (_that) {
case _TaxpayerUpdate():
return $default(_that.displayName);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'display_name')  String displayName)?  $default,) {final _that = this;
switch (_that) {
case _TaxpayerUpdate() when $default != null:
return $default(_that.displayName);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TaxpayerUpdate implements TaxpayerUpdate {
  const _TaxpayerUpdate({@JsonKey(name: 'display_name') required this.displayName});
  factory _TaxpayerUpdate.fromJson(Map<String, dynamic> json) => _$TaxpayerUpdateFromJson(json);

@override@JsonKey(name: 'display_name') final  String displayName;

/// Create a copy of TaxpayerUpdate
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TaxpayerUpdateCopyWith<_TaxpayerUpdate> get copyWith => __$TaxpayerUpdateCopyWithImpl<_TaxpayerUpdate>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TaxpayerUpdateToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TaxpayerUpdate&&(identical(other.displayName, displayName) || other.displayName == displayName));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,displayName);
}

@override
String toString() {
    return 'TaxpayerUpdate(displayName: $displayName)';
}


}

/// @nodoc
abstract mixin class _$TaxpayerUpdateCopyWith<$Res> implements $TaxpayerUpdateCopyWith<$Res> {
  factory _$TaxpayerUpdateCopyWith(_TaxpayerUpdate value, $Res Function(_TaxpayerUpdate) _then) = __$TaxpayerUpdateCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'display_name') String displayName
});




}
/// @nodoc
class __$TaxpayerUpdateCopyWithImpl<$Res>
    implements _$TaxpayerUpdateCopyWith<$Res> {
  __$TaxpayerUpdateCopyWithImpl(this._self, this._then);

  final _TaxpayerUpdate _self;
  final $Res Function(_TaxpayerUpdate) _then;

/// Create a copy of TaxpayerUpdate
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? displayName = null,}) {
  return _then(_TaxpayerUpdate(
displayName: null == displayName ? _self.displayName : displayName // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}

// dart format on
