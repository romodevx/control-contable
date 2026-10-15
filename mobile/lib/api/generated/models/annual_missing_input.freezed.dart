// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'annual_missing_input.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$AnnualMissingInput {

 String get key;/// Spanish label for the CTA.
 String get label; AnnualInputTarget get target;@JsonKey(includeIfNull: true, name: 'regime_code') RegimeCode? get regimeCode; List<Period> get months;
/// Create a copy of AnnualMissingInput
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$AnnualMissingInputCopyWith<AnnualMissingInput> get copyWith => _$AnnualMissingInputCopyWithImpl<AnnualMissingInput>(this as AnnualMissingInput, _$identity);

  /// Serializes this AnnualMissingInput to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as AnnualMissingInput;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is AnnualMissingInput&&(identical(other.key, _this.key) || other.key == _this.key)&&(identical(other.label, _this.label) || other.label == _this.label)&&(identical(other.target, _this.target) || other.target == _this.target)&&(identical(other.regimeCode, _this.regimeCode) || other.regimeCode == _this.regimeCode)&&const DeepCollectionEquality().equals(other.months, _this.months));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as AnnualMissingInput;
  return Object.hash(runtimeType,_this.key,_this.label,_this.target,_this.regimeCode,const DeepCollectionEquality().hash(_this.months));
}

@override
String toString() {
  final _this = this as AnnualMissingInput;
  return 'AnnualMissingInput(key: ${_this.key}, label: ${_this.label}, target: ${_this.target}, regimeCode: ${_this.regimeCode}, months: ${_this.months})';
}


}

/// @nodoc
abstract mixin class $AnnualMissingInputCopyWith<$Res>  {
  factory $AnnualMissingInputCopyWith(AnnualMissingInput value, $Res Function(AnnualMissingInput) _then) = _$AnnualMissingInputCopyWithImpl;
@useResult
$Res call({
 String key, String label, AnnualInputTarget target,@JsonKey(includeIfNull: true, name: 'regime_code') RegimeCode? regimeCode, List<Period> months
});




}
/// @nodoc
class _$AnnualMissingInputCopyWithImpl<$Res>
    implements $AnnualMissingInputCopyWith<$Res> {
  _$AnnualMissingInputCopyWithImpl(this._self, this._then);

  final AnnualMissingInput _self;
  final $Res Function(AnnualMissingInput) _then;

/// Create a copy of AnnualMissingInput
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? key = null,Object? label = null,Object? target = null,Object? regimeCode = freezed,Object? months = null,}) {
  return _then(AnnualMissingInput(
key: null == key ? _self.key : key // ignore: cast_nullable_to_non_nullable
as String,label: null == label ? _self.label : label // ignore: cast_nullable_to_non_nullable
as String,target: null == target ? _self.target : target // ignore: cast_nullable_to_non_nullable
as AnnualInputTarget,regimeCode: freezed == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode?,months: null == months ? _self.months : months // ignore: cast_nullable_to_non_nullable
as List<Period>,
  ));
}

}


/// Adds pattern-matching-related methods to [AnnualMissingInput].
extension AnnualMissingInputPatterns on AnnualMissingInput {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _AnnualMissingInput value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _AnnualMissingInput() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _AnnualMissingInput value)  $default,){
final _that = this;
switch (_that) {
case _AnnualMissingInput():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _AnnualMissingInput value)?  $default,){
final _that = this;
switch (_that) {
case _AnnualMissingInput() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String key,  String label,  AnnualInputTarget target, @JsonKey(includeIfNull: true, name: 'regime_code')  RegimeCode? regimeCode,  List<Period> months)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _AnnualMissingInput() when $default != null:
return $default(_that.key,_that.label,_that.target,_that.regimeCode,_that.months);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String key,  String label,  AnnualInputTarget target, @JsonKey(includeIfNull: true, name: 'regime_code')  RegimeCode? regimeCode,  List<Period> months)  $default,) {final _that = this;
switch (_that) {
case _AnnualMissingInput():
return $default(_that.key,_that.label,_that.target,_that.regimeCode,_that.months);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String key,  String label,  AnnualInputTarget target, @JsonKey(includeIfNull: true, name: 'regime_code')  RegimeCode? regimeCode,  List<Period> months)?  $default,) {final _that = this;
switch (_that) {
case _AnnualMissingInput() when $default != null:
return $default(_that.key,_that.label,_that.target,_that.regimeCode,_that.months);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _AnnualMissingInput implements AnnualMissingInput {
  const _AnnualMissingInput({required this.key, required this.label, required this.target, @JsonKey(includeIfNull: true, name: 'regime_code') required this.regimeCode, required  List<Period> months}): _months = months;
  factory _AnnualMissingInput.fromJson(Map<String, dynamic> json) => _$AnnualMissingInputFromJson(json);

@override final  String key;
/// Spanish label for the CTA.
@override final  String label;
@override final  AnnualInputTarget target;
@override@JsonKey(includeIfNull: true, name: 'regime_code') final  RegimeCode? regimeCode;
 final  List<Period> _months;
@override List<Period> get months {
  if (_months is EqualUnmodifiableListView) return _months;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_months);
}


/// Create a copy of AnnualMissingInput
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$AnnualMissingInputCopyWith<_AnnualMissingInput> get copyWith => __$AnnualMissingInputCopyWithImpl<_AnnualMissingInput>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$AnnualMissingInputToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _AnnualMissingInput&&(identical(other.key, key) || other.key == key)&&(identical(other.label, label) || other.label == label)&&(identical(other.target, target) || other.target == target)&&(identical(other.regimeCode, regimeCode) || other.regimeCode == regimeCode)&&const DeepCollectionEquality().equals(other.months, _months));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,key,label,target,regimeCode,const DeepCollectionEquality().hash(_months));
}

@override
String toString() {
    return 'AnnualMissingInput(key: $key, label: $label, target: $target, regimeCode: $regimeCode, months: $months)';
}


}

/// @nodoc
abstract mixin class _$AnnualMissingInputCopyWith<$Res> implements $AnnualMissingInputCopyWith<$Res> {
  factory _$AnnualMissingInputCopyWith(_AnnualMissingInput value, $Res Function(_AnnualMissingInput) _then) = __$AnnualMissingInputCopyWithImpl;
@override @useResult
$Res call({
 String key, String label, AnnualInputTarget target,@JsonKey(includeIfNull: true, name: 'regime_code') RegimeCode? regimeCode, List<Period> months
});




}
/// @nodoc
class __$AnnualMissingInputCopyWithImpl<$Res>
    implements _$AnnualMissingInputCopyWith<$Res> {
  __$AnnualMissingInputCopyWithImpl(this._self, this._then);

  final _AnnualMissingInput _self;
  final $Res Function(_AnnualMissingInput) _then;

/// Create a copy of AnnualMissingInput
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? key = null,Object? label = null,Object? target = null,Object? regimeCode = freezed,Object? months = null,}) {
  return _then(_AnnualMissingInput(
key: null == key ? _self.key : key // ignore: cast_nullable_to_non_nullable
as String,label: null == label ? _self.label : label // ignore: cast_nullable_to_non_nullable
as String,target: null == target ? _self.target : target // ignore: cast_nullable_to_non_nullable
as AnnualInputTarget,regimeCode: freezed == regimeCode ? _self.regimeCode : regimeCode // ignore: cast_nullable_to_non_nullable
as RegimeCode?,months: null == months ? _self._months : months // ignore: cast_nullable_to_non_nullable
as List<Period>,
  ));
}


}

// dart format on
