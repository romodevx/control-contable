// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'tax_inputs_update.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$TaxInputsUpdate {

 List<TaxInputValue> get values;
/// Create a copy of TaxInputsUpdate
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TaxInputsUpdateCopyWith<TaxInputsUpdate> get copyWith => _$TaxInputsUpdateCopyWithImpl<TaxInputsUpdate>(this as TaxInputsUpdate, _$identity);

  /// Serializes this TaxInputsUpdate to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as TaxInputsUpdate;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TaxInputsUpdate&&const DeepCollectionEquality().equals(other.values, _this.values));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TaxInputsUpdate;
  return Object.hash(runtimeType,const DeepCollectionEquality().hash(_this.values));
}

@override
String toString() {
  final _this = this as TaxInputsUpdate;
  return 'TaxInputsUpdate(values: ${_this.values})';
}


}

/// @nodoc
abstract mixin class $TaxInputsUpdateCopyWith<$Res>  {
  factory $TaxInputsUpdateCopyWith(TaxInputsUpdate value, $Res Function(TaxInputsUpdate) _then) = _$TaxInputsUpdateCopyWithImpl;
@useResult
$Res call({
 List<TaxInputValue> values
});




}
/// @nodoc
class _$TaxInputsUpdateCopyWithImpl<$Res>
    implements $TaxInputsUpdateCopyWith<$Res> {
  _$TaxInputsUpdateCopyWithImpl(this._self, this._then);

  final TaxInputsUpdate _self;
  final $Res Function(TaxInputsUpdate) _then;

/// Create a copy of TaxInputsUpdate
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? values = null,}) {
  return _then(TaxInputsUpdate(
values: null == values ? _self.values : values // ignore: cast_nullable_to_non_nullable
as List<TaxInputValue>,
  ));
}

}


/// Adds pattern-matching-related methods to [TaxInputsUpdate].
extension TaxInputsUpdatePatterns on TaxInputsUpdate {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _TaxInputsUpdate value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _TaxInputsUpdate() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _TaxInputsUpdate value)  $default,){
final _that = this;
switch (_that) {
case _TaxInputsUpdate():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _TaxInputsUpdate value)?  $default,){
final _that = this;
switch (_that) {
case _TaxInputsUpdate() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( List<TaxInputValue> values)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TaxInputsUpdate() when $default != null:
return $default(_that.values);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( List<TaxInputValue> values)  $default,) {final _that = this;
switch (_that) {
case _TaxInputsUpdate():
return $default(_that.values);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( List<TaxInputValue> values)?  $default,) {final _that = this;
switch (_that) {
case _TaxInputsUpdate() when $default != null:
return $default(_that.values);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TaxInputsUpdate implements TaxInputsUpdate {
  const _TaxInputsUpdate({required  List<TaxInputValue> values}): _values = values;
  factory _TaxInputsUpdate.fromJson(Map<String, dynamic> json) => _$TaxInputsUpdateFromJson(json);

 final  List<TaxInputValue> _values;
@override List<TaxInputValue> get values {
  if (_values is EqualUnmodifiableListView) return _values;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_values);
}


/// Create a copy of TaxInputsUpdate
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TaxInputsUpdateCopyWith<_TaxInputsUpdate> get copyWith => __$TaxInputsUpdateCopyWithImpl<_TaxInputsUpdate>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TaxInputsUpdateToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TaxInputsUpdate&&const DeepCollectionEquality().equals(other.values, _values));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,const DeepCollectionEquality().hash(_values));
}

@override
String toString() {
    return 'TaxInputsUpdate(values: $values)';
}


}

/// @nodoc
abstract mixin class _$TaxInputsUpdateCopyWith<$Res> implements $TaxInputsUpdateCopyWith<$Res> {
  factory _$TaxInputsUpdateCopyWith(_TaxInputsUpdate value, $Res Function(_TaxInputsUpdate) _then) = __$TaxInputsUpdateCopyWithImpl;
@override @useResult
$Res call({
 List<TaxInputValue> values
});




}
/// @nodoc
class __$TaxInputsUpdateCopyWithImpl<$Res>
    implements _$TaxInputsUpdateCopyWith<$Res> {
  __$TaxInputsUpdateCopyWithImpl(this._self, this._then);

  final _TaxInputsUpdate _self;
  final $Res Function(_TaxInputsUpdate) _then;

/// Create a copy of TaxInputsUpdate
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? values = null,}) {
  return _then(_TaxInputsUpdate(
values: null == values ? _self._values : values // ignore: cast_nullable_to_non_nullable
as List<TaxInputValue>,
  ));
}


}

// dart format on
