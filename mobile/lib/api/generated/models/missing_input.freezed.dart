// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'missing_input.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$MissingInput {

 String get key;/// Spanish label (`TaxInputDef.label`).
 String get label;/// Only for `money_monthly` keys.
@JsonKey(includeIfNull: false) List<Period>? get months;
/// Create a copy of MissingInput
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$MissingInputCopyWith<MissingInput> get copyWith => _$MissingInputCopyWithImpl<MissingInput>(this as MissingInput, _$identity);

  /// Serializes this MissingInput to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as MissingInput;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is MissingInput&&(identical(other.key, _this.key) || other.key == _this.key)&&(identical(other.label, _this.label) || other.label == _this.label)&&const DeepCollectionEquality().equals(other.months, _this.months));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as MissingInput;
  return Object.hash(runtimeType,_this.key,_this.label,const DeepCollectionEquality().hash(_this.months));
}

@override
String toString() {
  final _this = this as MissingInput;
  return 'MissingInput(key: ${_this.key}, label: ${_this.label}, months: ${_this.months})';
}


}

/// @nodoc
abstract mixin class $MissingInputCopyWith<$Res>  {
  factory $MissingInputCopyWith(MissingInput value, $Res Function(MissingInput) _then) = _$MissingInputCopyWithImpl;
@useResult
$Res call({
 String key, String label,@JsonKey(includeIfNull: false) List<Period>? months
});




}
/// @nodoc
class _$MissingInputCopyWithImpl<$Res>
    implements $MissingInputCopyWith<$Res> {
  _$MissingInputCopyWithImpl(this._self, this._then);

  final MissingInput _self;
  final $Res Function(MissingInput) _then;

/// Create a copy of MissingInput
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? key = null,Object? label = null,Object? months = freezed,}) {
  return _then(MissingInput(
key: null == key ? _self.key : key // ignore: cast_nullable_to_non_nullable
as String,label: null == label ? _self.label : label // ignore: cast_nullable_to_non_nullable
as String,months: freezed == months ? _self.months : months // ignore: cast_nullable_to_non_nullable
as List<Period>?,
  ));
}

}


/// Adds pattern-matching-related methods to [MissingInput].
extension MissingInputPatterns on MissingInput {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _MissingInput value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _MissingInput() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _MissingInput value)  $default,){
final _that = this;
switch (_that) {
case _MissingInput():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _MissingInput value)?  $default,){
final _that = this;
switch (_that) {
case _MissingInput() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String key,  String label, @JsonKey(includeIfNull: false)  List<Period>? months)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _MissingInput() when $default != null:
return $default(_that.key,_that.label,_that.months);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String key,  String label, @JsonKey(includeIfNull: false)  List<Period>? months)  $default,) {final _that = this;
switch (_that) {
case _MissingInput():
return $default(_that.key,_that.label,_that.months);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String key,  String label, @JsonKey(includeIfNull: false)  List<Period>? months)?  $default,) {final _that = this;
switch (_that) {
case _MissingInput() when $default != null:
return $default(_that.key,_that.label,_that.months);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _MissingInput implements MissingInput {
  const _MissingInput({required this.key, required this.label, @JsonKey(includeIfNull: false)  List<Period>? months}): _months = months;
  factory _MissingInput.fromJson(Map<String, dynamic> json) => _$MissingInputFromJson(json);

@override final  String key;
/// Spanish label (`TaxInputDef.label`).
@override final  String label;
/// Only for `money_monthly` keys.
 final  List<Period>? _months;
/// Only for `money_monthly` keys.
@override@JsonKey(includeIfNull: false) List<Period>? get months {
  final value = _months;
  if (value == null) return null;
  if (_months is EqualUnmodifiableListView) return _months;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(value);
}


/// Create a copy of MissingInput
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$MissingInputCopyWith<_MissingInput> get copyWith => __$MissingInputCopyWithImpl<_MissingInput>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$MissingInputToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _MissingInput&&(identical(other.key, key) || other.key == key)&&(identical(other.label, label) || other.label == label)&&const DeepCollectionEquality().equals(other.months, _months));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,key,label,const DeepCollectionEquality().hash(_months));
}

@override
String toString() {
    return 'MissingInput(key: $key, label: $label, months: $months)';
}


}

/// @nodoc
abstract mixin class _$MissingInputCopyWith<$Res> implements $MissingInputCopyWith<$Res> {
  factory _$MissingInputCopyWith(_MissingInput value, $Res Function(_MissingInput) _then) = __$MissingInputCopyWithImpl;
@override @useResult
$Res call({
 String key, String label,@JsonKey(includeIfNull: false) List<Period>? months
});




}
/// @nodoc
class __$MissingInputCopyWithImpl<$Res>
    implements _$MissingInputCopyWith<$Res> {
  __$MissingInputCopyWithImpl(this._self, this._then);

  final _MissingInput _self;
  final $Res Function(_MissingInput) _then;

/// Create a copy of MissingInput
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? key = null,Object? label = null,Object? months = freezed,}) {
  return _then(_MissingInput(
key: null == key ? _self.key : key // ignore: cast_nullable_to_non_nullable
as String,label: null == label ? _self.label : label // ignore: cast_nullable_to_non_nullable
as String,months: freezed == months ? _self._months : months // ignore: cast_nullable_to_non_nullable
as List<Period>?,
  ));
}


}

// dart format on
