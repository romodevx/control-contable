// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'regime.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$Regime {

 RegimeCode get code; String get name;@JsonKey(name: 'person_type') PersonType get personType; Treatment get treatment;@JsonKey(name: 'option_defs') List<RegimeOptionDef> get optionDefs;/// Tax-input keys the strategy needs (see TaxInputs).
@JsonKey(name: 'required_inputs') List<String> get requiredInputs;/// Plain-language reason for partial/informational treatment.
 String? get explanation;
/// Create a copy of Regime
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$RegimeCopyWith<Regime> get copyWith => _$RegimeCopyWithImpl<Regime>(this as Regime, _$identity);

  /// Serializes this Regime to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as Regime;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is Regime&&(identical(other.code, _this.code) || other.code == _this.code)&&(identical(other.name, _this.name) || other.name == _this.name)&&(identical(other.personType, _this.personType) || other.personType == _this.personType)&&(identical(other.treatment, _this.treatment) || other.treatment == _this.treatment)&&const DeepCollectionEquality().equals(other.optionDefs, _this.optionDefs)&&const DeepCollectionEquality().equals(other.requiredInputs, _this.requiredInputs)&&(identical(other.explanation, _this.explanation) || other.explanation == _this.explanation));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as Regime;
  return Object.hash(runtimeType,_this.code,_this.name,_this.personType,_this.treatment,const DeepCollectionEquality().hash(_this.optionDefs),const DeepCollectionEquality().hash(_this.requiredInputs),_this.explanation);
}

@override
String toString() {
  final _this = this as Regime;
  return 'Regime(code: ${_this.code}, name: ${_this.name}, personType: ${_this.personType}, treatment: ${_this.treatment}, optionDefs: ${_this.optionDefs}, requiredInputs: ${_this.requiredInputs}, explanation: ${_this.explanation})';
}


}

/// @nodoc
abstract mixin class $RegimeCopyWith<$Res>  {
  factory $RegimeCopyWith(Regime value, $Res Function(Regime) _then) = _$RegimeCopyWithImpl;
@useResult
$Res call({
 RegimeCode code, String name,@JsonKey(name: 'person_type') PersonType personType, Treatment treatment,@JsonKey(name: 'option_defs') List<RegimeOptionDef> optionDefs,@JsonKey(name: 'required_inputs') List<String> requiredInputs, String? explanation
});




}
/// @nodoc
class _$RegimeCopyWithImpl<$Res>
    implements $RegimeCopyWith<$Res> {
  _$RegimeCopyWithImpl(this._self, this._then);

  final Regime _self;
  final $Res Function(Regime) _then;

/// Create a copy of Regime
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? code = null,Object? name = null,Object? personType = null,Object? treatment = null,Object? optionDefs = null,Object? requiredInputs = null,Object? explanation = freezed,}) {
  return _then(Regime(
code: null == code ? _self.code : code // ignore: cast_nullable_to_non_nullable
as RegimeCode,name: null == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as String,personType: null == personType ? _self.personType : personType // ignore: cast_nullable_to_non_nullable
as PersonType,treatment: null == treatment ? _self.treatment : treatment // ignore: cast_nullable_to_non_nullable
as Treatment,optionDefs: null == optionDefs ? _self.optionDefs : optionDefs // ignore: cast_nullable_to_non_nullable
as List<RegimeOptionDef>,requiredInputs: null == requiredInputs ? _self.requiredInputs : requiredInputs // ignore: cast_nullable_to_non_nullable
as List<String>,explanation: freezed == explanation ? _self.explanation : explanation // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}

}


/// Adds pattern-matching-related methods to [Regime].
extension RegimePatterns on Regime {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _Regime value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _Regime() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _Regime value)  $default,){
final _that = this;
switch (_that) {
case _Regime():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _Regime value)?  $default,){
final _that = this;
switch (_that) {
case _Regime() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( RegimeCode code,  String name, @JsonKey(name: 'person_type')  PersonType personType,  Treatment treatment, @JsonKey(name: 'option_defs')  List<RegimeOptionDef> optionDefs, @JsonKey(name: 'required_inputs')  List<String> requiredInputs,  String? explanation)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _Regime() when $default != null:
return $default(_that.code,_that.name,_that.personType,_that.treatment,_that.optionDefs,_that.requiredInputs,_that.explanation);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( RegimeCode code,  String name, @JsonKey(name: 'person_type')  PersonType personType,  Treatment treatment, @JsonKey(name: 'option_defs')  List<RegimeOptionDef> optionDefs, @JsonKey(name: 'required_inputs')  List<String> requiredInputs,  String? explanation)  $default,) {final _that = this;
switch (_that) {
case _Regime():
return $default(_that.code,_that.name,_that.personType,_that.treatment,_that.optionDefs,_that.requiredInputs,_that.explanation);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( RegimeCode code,  String name, @JsonKey(name: 'person_type')  PersonType personType,  Treatment treatment, @JsonKey(name: 'option_defs')  List<RegimeOptionDef> optionDefs, @JsonKey(name: 'required_inputs')  List<String> requiredInputs,  String? explanation)?  $default,) {final _that = this;
switch (_that) {
case _Regime() when $default != null:
return $default(_that.code,_that.name,_that.personType,_that.treatment,_that.optionDefs,_that.requiredInputs,_that.explanation);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _Regime implements Regime {
  const _Regime({required this.code, required this.name, @JsonKey(name: 'person_type') required this.personType, required this.treatment, @JsonKey(name: 'option_defs') required  List<RegimeOptionDef> optionDefs, @JsonKey(name: 'required_inputs') required  List<String> requiredInputs, this.explanation}): _optionDefs = optionDefs,_requiredInputs = requiredInputs;
  factory _Regime.fromJson(Map<String, dynamic> json) => _$RegimeFromJson(json);

@override final  RegimeCode code;
@override final  String name;
@override@JsonKey(name: 'person_type') final  PersonType personType;
@override final  Treatment treatment;
 final  List<RegimeOptionDef> _optionDefs;
@override@JsonKey(name: 'option_defs') List<RegimeOptionDef> get optionDefs {
  if (_optionDefs is EqualUnmodifiableListView) return _optionDefs;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_optionDefs);
}

/// Tax-input keys the strategy needs (see TaxInputs).
 final  List<String> _requiredInputs;
/// Tax-input keys the strategy needs (see TaxInputs).
@override@JsonKey(name: 'required_inputs') List<String> get requiredInputs {
  if (_requiredInputs is EqualUnmodifiableListView) return _requiredInputs;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_requiredInputs);
}

/// Plain-language reason for partial/informational treatment.
@override final  String? explanation;

/// Create a copy of Regime
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$RegimeCopyWith<_Regime> get copyWith => __$RegimeCopyWithImpl<_Regime>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$RegimeToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _Regime&&(identical(other.code, code) || other.code == code)&&(identical(other.name, name) || other.name == name)&&(identical(other.personType, personType) || other.personType == personType)&&(identical(other.treatment, treatment) || other.treatment == treatment)&&const DeepCollectionEquality().equals(other.optionDefs, _optionDefs)&&const DeepCollectionEquality().equals(other.requiredInputs, _requiredInputs)&&(identical(other.explanation, explanation) || other.explanation == explanation));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,code,name,personType,treatment,const DeepCollectionEquality().hash(_optionDefs),const DeepCollectionEquality().hash(_requiredInputs),explanation);
}

@override
String toString() {
    return 'Regime(code: $code, name: $name, personType: $personType, treatment: $treatment, optionDefs: $optionDefs, requiredInputs: $requiredInputs, explanation: $explanation)';
}


}

/// @nodoc
abstract mixin class _$RegimeCopyWith<$Res> implements $RegimeCopyWith<$Res> {
  factory _$RegimeCopyWith(_Regime value, $Res Function(_Regime) _then) = __$RegimeCopyWithImpl;
@override @useResult
$Res call({
 RegimeCode code, String name,@JsonKey(name: 'person_type') PersonType personType, Treatment treatment,@JsonKey(name: 'option_defs') List<RegimeOptionDef> optionDefs,@JsonKey(name: 'required_inputs') List<String> requiredInputs, String? explanation
});




}
/// @nodoc
class __$RegimeCopyWithImpl<$Res>
    implements _$RegimeCopyWith<$Res> {
  __$RegimeCopyWithImpl(this._self, this._then);

  final _Regime _self;
  final $Res Function(_Regime) _then;

/// Create a copy of Regime
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? code = null,Object? name = null,Object? personType = null,Object? treatment = null,Object? optionDefs = null,Object? requiredInputs = null,Object? explanation = freezed,}) {
  return _then(_Regime(
code: null == code ? _self.code : code // ignore: cast_nullable_to_non_nullable
as RegimeCode,name: null == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as String,personType: null == personType ? _self.personType : personType // ignore: cast_nullable_to_non_nullable
as PersonType,treatment: null == treatment ? _self.treatment : treatment // ignore: cast_nullable_to_non_nullable
as Treatment,optionDefs: null == optionDefs ? _self._optionDefs : optionDefs // ignore: cast_nullable_to_non_nullable
as List<RegimeOptionDef>,requiredInputs: null == requiredInputs ? _self._requiredInputs : requiredInputs // ignore: cast_nullable_to_non_nullable
as List<String>,explanation: freezed == explanation ? _self.explanation : explanation // ignore: cast_nullable_to_non_nullable
as String?,
  ));
}


}

// dart format on
