// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'annual_inputs_update.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$AnnualInputsUpdate {

@JsonKey(includeIfNull: true, name: 'prior_year_accumulable_income_cents') NonNegativeCents? get priorYearAccumulableIncomeCents;@JsonKey(name: 'other_exempt_income_cents') NonNegativeCents get otherExemptIncomeCents;@JsonKey(name: 'told_employer_will_file') bool get toldEmployerWillFile;
/// Create a copy of AnnualInputsUpdate
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$AnnualInputsUpdateCopyWith<AnnualInputsUpdate> get copyWith => _$AnnualInputsUpdateCopyWithImpl<AnnualInputsUpdate>(this as AnnualInputsUpdate, _$identity);

  /// Serializes this AnnualInputsUpdate to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as AnnualInputsUpdate;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is AnnualInputsUpdate&&(identical(other.priorYearAccumulableIncomeCents, _this.priorYearAccumulableIncomeCents) || other.priorYearAccumulableIncomeCents == _this.priorYearAccumulableIncomeCents)&&(identical(other.otherExemptIncomeCents, _this.otherExemptIncomeCents) || other.otherExemptIncomeCents == _this.otherExemptIncomeCents)&&(identical(other.toldEmployerWillFile, _this.toldEmployerWillFile) || other.toldEmployerWillFile == _this.toldEmployerWillFile));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as AnnualInputsUpdate;
  return Object.hash(runtimeType,_this.priorYearAccumulableIncomeCents,_this.otherExemptIncomeCents,_this.toldEmployerWillFile);
}

@override
String toString() {
  final _this = this as AnnualInputsUpdate;
  return 'AnnualInputsUpdate(priorYearAccumulableIncomeCents: ${_this.priorYearAccumulableIncomeCents}, otherExemptIncomeCents: ${_this.otherExemptIncomeCents}, toldEmployerWillFile: ${_this.toldEmployerWillFile})';
}


}

/// @nodoc
abstract mixin class $AnnualInputsUpdateCopyWith<$Res>  {
  factory $AnnualInputsUpdateCopyWith(AnnualInputsUpdate value, $Res Function(AnnualInputsUpdate) _then) = _$AnnualInputsUpdateCopyWithImpl;
@useResult
$Res call({
@JsonKey(includeIfNull: true, name: 'prior_year_accumulable_income_cents') NonNegativeCents? priorYearAccumulableIncomeCents,@JsonKey(name: 'other_exempt_income_cents') NonNegativeCents otherExemptIncomeCents,@JsonKey(name: 'told_employer_will_file') bool toldEmployerWillFile
});




}
/// @nodoc
class _$AnnualInputsUpdateCopyWithImpl<$Res>
    implements $AnnualInputsUpdateCopyWith<$Res> {
  _$AnnualInputsUpdateCopyWithImpl(this._self, this._then);

  final AnnualInputsUpdate _self;
  final $Res Function(AnnualInputsUpdate) _then;

/// Create a copy of AnnualInputsUpdate
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? priorYearAccumulableIncomeCents = freezed,Object? otherExemptIncomeCents = null,Object? toldEmployerWillFile = null,}) {
  return _then(AnnualInputsUpdate(
priorYearAccumulableIncomeCents: freezed == priorYearAccumulableIncomeCents ? _self.priorYearAccumulableIncomeCents : priorYearAccumulableIncomeCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,otherExemptIncomeCents: null == otherExemptIncomeCents ? _self.otherExemptIncomeCents : otherExemptIncomeCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,toldEmployerWillFile: null == toldEmployerWillFile ? _self.toldEmployerWillFile : toldEmployerWillFile // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}

}


/// Adds pattern-matching-related methods to [AnnualInputsUpdate].
extension AnnualInputsUpdatePatterns on AnnualInputsUpdate {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _AnnualInputsUpdate value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _AnnualInputsUpdate() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _AnnualInputsUpdate value)  $default,){
final _that = this;
switch (_that) {
case _AnnualInputsUpdate():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _AnnualInputsUpdate value)?  $default,){
final _that = this;
switch (_that) {
case _AnnualInputsUpdate() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(includeIfNull: true, name: 'prior_year_accumulable_income_cents')  NonNegativeCents? priorYearAccumulableIncomeCents, @JsonKey(name: 'other_exempt_income_cents')  NonNegativeCents otherExemptIncomeCents, @JsonKey(name: 'told_employer_will_file')  bool toldEmployerWillFile)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _AnnualInputsUpdate() when $default != null:
return $default(_that.priorYearAccumulableIncomeCents,_that.otherExemptIncomeCents,_that.toldEmployerWillFile);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(includeIfNull: true, name: 'prior_year_accumulable_income_cents')  NonNegativeCents? priorYearAccumulableIncomeCents, @JsonKey(name: 'other_exempt_income_cents')  NonNegativeCents otherExemptIncomeCents, @JsonKey(name: 'told_employer_will_file')  bool toldEmployerWillFile)  $default,) {final _that = this;
switch (_that) {
case _AnnualInputsUpdate():
return $default(_that.priorYearAccumulableIncomeCents,_that.otherExemptIncomeCents,_that.toldEmployerWillFile);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(includeIfNull: true, name: 'prior_year_accumulable_income_cents')  NonNegativeCents? priorYearAccumulableIncomeCents, @JsonKey(name: 'other_exempt_income_cents')  NonNegativeCents otherExemptIncomeCents, @JsonKey(name: 'told_employer_will_file')  bool toldEmployerWillFile)?  $default,) {final _that = this;
switch (_that) {
case _AnnualInputsUpdate() when $default != null:
return $default(_that.priorYearAccumulableIncomeCents,_that.otherExemptIncomeCents,_that.toldEmployerWillFile);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _AnnualInputsUpdate implements AnnualInputsUpdate {
  const _AnnualInputsUpdate({@JsonKey(includeIfNull: true, name: 'prior_year_accumulable_income_cents') required this.priorYearAccumulableIncomeCents, @JsonKey(name: 'other_exempt_income_cents') required this.otherExemptIncomeCents, @JsonKey(name: 'told_employer_will_file') required this.toldEmployerWillFile});
  factory _AnnualInputsUpdate.fromJson(Map<String, dynamic> json) => _$AnnualInputsUpdateFromJson(json);

@override@JsonKey(includeIfNull: true, name: 'prior_year_accumulable_income_cents') final  NonNegativeCents? priorYearAccumulableIncomeCents;
@override@JsonKey(name: 'other_exempt_income_cents') final  NonNegativeCents otherExemptIncomeCents;
@override@JsonKey(name: 'told_employer_will_file') final  bool toldEmployerWillFile;

/// Create a copy of AnnualInputsUpdate
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$AnnualInputsUpdateCopyWith<_AnnualInputsUpdate> get copyWith => __$AnnualInputsUpdateCopyWithImpl<_AnnualInputsUpdate>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$AnnualInputsUpdateToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _AnnualInputsUpdate&&(identical(other.priorYearAccumulableIncomeCents, priorYearAccumulableIncomeCents) || other.priorYearAccumulableIncomeCents == priorYearAccumulableIncomeCents)&&(identical(other.otherExemptIncomeCents, otherExemptIncomeCents) || other.otherExemptIncomeCents == otherExemptIncomeCents)&&(identical(other.toldEmployerWillFile, toldEmployerWillFile) || other.toldEmployerWillFile == toldEmployerWillFile));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,priorYearAccumulableIncomeCents,otherExemptIncomeCents,toldEmployerWillFile);
}

@override
String toString() {
    return 'AnnualInputsUpdate(priorYearAccumulableIncomeCents: $priorYearAccumulableIncomeCents, otherExemptIncomeCents: $otherExemptIncomeCents, toldEmployerWillFile: $toldEmployerWillFile)';
}


}

/// @nodoc
abstract mixin class _$AnnualInputsUpdateCopyWith<$Res> implements $AnnualInputsUpdateCopyWith<$Res> {
  factory _$AnnualInputsUpdateCopyWith(_AnnualInputsUpdate value, $Res Function(_AnnualInputsUpdate) _then) = __$AnnualInputsUpdateCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(includeIfNull: true, name: 'prior_year_accumulable_income_cents') NonNegativeCents? priorYearAccumulableIncomeCents,@JsonKey(name: 'other_exempt_income_cents') NonNegativeCents otherExemptIncomeCents,@JsonKey(name: 'told_employer_will_file') bool toldEmployerWillFile
});




}
/// @nodoc
class __$AnnualInputsUpdateCopyWithImpl<$Res>
    implements _$AnnualInputsUpdateCopyWith<$Res> {
  __$AnnualInputsUpdateCopyWithImpl(this._self, this._then);

  final _AnnualInputsUpdate _self;
  final $Res Function(_AnnualInputsUpdate) _then;

/// Create a copy of AnnualInputsUpdate
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? priorYearAccumulableIncomeCents = freezed,Object? otherExemptIncomeCents = null,Object? toldEmployerWillFile = null,}) {
  return _then(_AnnualInputsUpdate(
priorYearAccumulableIncomeCents: freezed == priorYearAccumulableIncomeCents ? _self.priorYearAccumulableIncomeCents : priorYearAccumulableIncomeCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents?,otherExemptIncomeCents: null == otherExemptIncomeCents ? _self.otherExemptIncomeCents : otherExemptIncomeCents // ignore: cast_nullable_to_non_nullable
as NonNegativeCents,toldEmployerWillFile: null == toldEmployerWillFile ? _self.toldEmployerWillFile : toldEmployerWillFile // ignore: cast_nullable_to_non_nullable
as bool,
  ));
}


}

// dart format on
