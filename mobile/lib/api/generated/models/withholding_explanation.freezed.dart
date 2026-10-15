// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'withholding_explanation.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$WithholdingExplanation {

@JsonKey(name: 'rule_id') WithholdingRuleId get ruleId;/// One plain-words line for a matched rule, with its legal basis (es-MX), written by the server from the decision table. Safe to show; never an RFC, a name or a description.
 String get text;
/// Create a copy of WithholdingExplanation
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$WithholdingExplanationCopyWith<WithholdingExplanation> get copyWith => _$WithholdingExplanationCopyWithImpl<WithholdingExplanation>(this as WithholdingExplanation, _$identity);

  /// Serializes this WithholdingExplanation to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as WithholdingExplanation;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is WithholdingExplanation&&(identical(other.ruleId, _this.ruleId) || other.ruleId == _this.ruleId)&&(identical(other.text, _this.text) || other.text == _this.text));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as WithholdingExplanation;
  return Object.hash(runtimeType,_this.ruleId,_this.text);
}

@override
String toString() {
  final _this = this as WithholdingExplanation;
  return 'WithholdingExplanation(ruleId: ${_this.ruleId}, text: ${_this.text})';
}


}

/// @nodoc
abstract mixin class $WithholdingExplanationCopyWith<$Res>  {
  factory $WithholdingExplanationCopyWith(WithholdingExplanation value, $Res Function(WithholdingExplanation) _then) = _$WithholdingExplanationCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'rule_id') WithholdingRuleId ruleId, String text
});




}
/// @nodoc
class _$WithholdingExplanationCopyWithImpl<$Res>
    implements $WithholdingExplanationCopyWith<$Res> {
  _$WithholdingExplanationCopyWithImpl(this._self, this._then);

  final WithholdingExplanation _self;
  final $Res Function(WithholdingExplanation) _then;

/// Create a copy of WithholdingExplanation
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? ruleId = null,Object? text = null,}) {
  return _then(WithholdingExplanation(
ruleId: null == ruleId ? _self.ruleId : ruleId // ignore: cast_nullable_to_non_nullable
as WithholdingRuleId,text: null == text ? _self.text : text // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [WithholdingExplanation].
extension WithholdingExplanationPatterns on WithholdingExplanation {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _WithholdingExplanation value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _WithholdingExplanation() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _WithholdingExplanation value)  $default,){
final _that = this;
switch (_that) {
case _WithholdingExplanation():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _WithholdingExplanation value)?  $default,){
final _that = this;
switch (_that) {
case _WithholdingExplanation() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'rule_id')  WithholdingRuleId ruleId,  String text)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _WithholdingExplanation() when $default != null:
return $default(_that.ruleId,_that.text);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'rule_id')  WithholdingRuleId ruleId,  String text)  $default,) {final _that = this;
switch (_that) {
case _WithholdingExplanation():
return $default(_that.ruleId,_that.text);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'rule_id')  WithholdingRuleId ruleId,  String text)?  $default,) {final _that = this;
switch (_that) {
case _WithholdingExplanation() when $default != null:
return $default(_that.ruleId,_that.text);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _WithholdingExplanation implements WithholdingExplanation {
  const _WithholdingExplanation({@JsonKey(name: 'rule_id') required this.ruleId, required this.text});
  factory _WithholdingExplanation.fromJson(Map<String, dynamic> json) => _$WithholdingExplanationFromJson(json);

@override@JsonKey(name: 'rule_id') final  WithholdingRuleId ruleId;
/// One plain-words line for a matched rule, with its legal basis (es-MX), written by the server from the decision table. Safe to show; never an RFC, a name or a description.
@override final  String text;

/// Create a copy of WithholdingExplanation
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$WithholdingExplanationCopyWith<_WithholdingExplanation> get copyWith => __$WithholdingExplanationCopyWithImpl<_WithholdingExplanation>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$WithholdingExplanationToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _WithholdingExplanation&&(identical(other.ruleId, ruleId) || other.ruleId == ruleId)&&(identical(other.text, text) || other.text == text));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,ruleId,text);
}

@override
String toString() {
    return 'WithholdingExplanation(ruleId: $ruleId, text: $text)';
}


}

/// @nodoc
abstract mixin class _$WithholdingExplanationCopyWith<$Res> implements $WithholdingExplanationCopyWith<$Res> {
  factory _$WithholdingExplanationCopyWith(_WithholdingExplanation value, $Res Function(_WithholdingExplanation) _then) = __$WithholdingExplanationCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'rule_id') WithholdingRuleId ruleId, String text
});




}
/// @nodoc
class __$WithholdingExplanationCopyWithImpl<$Res>
    implements _$WithholdingExplanationCopyWith<$Res> {
  __$WithholdingExplanationCopyWithImpl(this._self, this._then);

  final _WithholdingExplanation _self;
  final $Res Function(_WithholdingExplanation) _then;

/// Create a copy of WithholdingExplanation
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? ruleId = null,Object? text = null,}) {
  return _then(_WithholdingExplanation(
ruleId: null == ruleId ? _self.ruleId : ruleId // ignore: cast_nullable_to_non_nullable
as WithholdingRuleId,text: null == text ? _self.text : text // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}

// dart format on
