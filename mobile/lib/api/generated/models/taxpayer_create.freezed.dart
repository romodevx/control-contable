// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'taxpayer_create.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$TaxpayerCreate {

@JsonKey(name: 'person_type') PersonType get personType; RfcInput get rfc;/// Trimmed and NFC-normalized; no invisible characters (control, format such as zero-width or bidi controls, line or paragraph separators; a zero-width joiner inside an emoji sequence is allowed): `400 validation_failed`, field code `invisible_characters`.
@JsonKey(name: 'display_name') String get displayName;/// At least one regime; each code at most once (`regimes.N.regime_code` / `duplicate`).
 List<TaxpayerRegimeCreate> get regimes;/// Optional client-generated id (idempotent retries).
@JsonKey(includeIfNull: false) Uuid? get id;
/// Create a copy of TaxpayerCreate
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TaxpayerCreateCopyWith<TaxpayerCreate> get copyWith => _$TaxpayerCreateCopyWithImpl<TaxpayerCreate>(this as TaxpayerCreate, _$identity);

  /// Serializes this TaxpayerCreate to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as TaxpayerCreate;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TaxpayerCreate&&(identical(other.personType, _this.personType) || other.personType == _this.personType)&&(identical(other.rfc, _this.rfc) || other.rfc == _this.rfc)&&(identical(other.displayName, _this.displayName) || other.displayName == _this.displayName)&&const DeepCollectionEquality().equals(other.regimes, _this.regimes)&&(identical(other.id, _this.id) || other.id == _this.id));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TaxpayerCreate;
  return Object.hash(runtimeType,_this.personType,_this.rfc,_this.displayName,const DeepCollectionEquality().hash(_this.regimes),_this.id);
}

@override
String toString() {
  final _this = this as TaxpayerCreate;
  return 'TaxpayerCreate(personType: ${_this.personType}, rfc: ${_this.rfc}, displayName: ${_this.displayName}, regimes: ${_this.regimes}, id: ${_this.id})';
}


}

/// @nodoc
abstract mixin class $TaxpayerCreateCopyWith<$Res>  {
  factory $TaxpayerCreateCopyWith(TaxpayerCreate value, $Res Function(TaxpayerCreate) _then) = _$TaxpayerCreateCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'person_type') PersonType personType, RfcInput rfc,@JsonKey(name: 'display_name') String displayName, List<TaxpayerRegimeCreate> regimes,@JsonKey(includeIfNull: false) Uuid? id
});




}
/// @nodoc
class _$TaxpayerCreateCopyWithImpl<$Res>
    implements $TaxpayerCreateCopyWith<$Res> {
  _$TaxpayerCreateCopyWithImpl(this._self, this._then);

  final TaxpayerCreate _self;
  final $Res Function(TaxpayerCreate) _then;

/// Create a copy of TaxpayerCreate
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? personType = null,Object? rfc = null,Object? displayName = null,Object? regimes = null,Object? id = freezed,}) {
  return _then(TaxpayerCreate(
personType: null == personType ? _self.personType : personType // ignore: cast_nullable_to_non_nullable
as PersonType,rfc: null == rfc ? _self.rfc : rfc // ignore: cast_nullable_to_non_nullable
as RfcInput,displayName: null == displayName ? _self.displayName : displayName // ignore: cast_nullable_to_non_nullable
as String,regimes: null == regimes ? _self.regimes : regimes // ignore: cast_nullable_to_non_nullable
as List<TaxpayerRegimeCreate>,id: freezed == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid?,
  ));
}

}


/// Adds pattern-matching-related methods to [TaxpayerCreate].
extension TaxpayerCreatePatterns on TaxpayerCreate {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _TaxpayerCreate value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _TaxpayerCreate() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _TaxpayerCreate value)  $default,){
final _that = this;
switch (_that) {
case _TaxpayerCreate():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _TaxpayerCreate value)?  $default,){
final _that = this;
switch (_that) {
case _TaxpayerCreate() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'person_type')  PersonType personType,  RfcInput rfc, @JsonKey(name: 'display_name')  String displayName,  List<TaxpayerRegimeCreate> regimes, @JsonKey(includeIfNull: false)  Uuid? id)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TaxpayerCreate() when $default != null:
return $default(_that.personType,_that.rfc,_that.displayName,_that.regimes,_that.id);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'person_type')  PersonType personType,  RfcInput rfc, @JsonKey(name: 'display_name')  String displayName,  List<TaxpayerRegimeCreate> regimes, @JsonKey(includeIfNull: false)  Uuid? id)  $default,) {final _that = this;
switch (_that) {
case _TaxpayerCreate():
return $default(_that.personType,_that.rfc,_that.displayName,_that.regimes,_that.id);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'person_type')  PersonType personType,  RfcInput rfc, @JsonKey(name: 'display_name')  String displayName,  List<TaxpayerRegimeCreate> regimes, @JsonKey(includeIfNull: false)  Uuid? id)?  $default,) {final _that = this;
switch (_that) {
case _TaxpayerCreate() when $default != null:
return $default(_that.personType,_that.rfc,_that.displayName,_that.regimes,_that.id);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TaxpayerCreate implements TaxpayerCreate {
  const _TaxpayerCreate({@JsonKey(name: 'person_type') required this.personType, required this.rfc, @JsonKey(name: 'display_name') required this.displayName, required  List<TaxpayerRegimeCreate> regimes, @JsonKey(includeIfNull: false) this.id}): _regimes = regimes;
  factory _TaxpayerCreate.fromJson(Map<String, dynamic> json) => _$TaxpayerCreateFromJson(json);

@override@JsonKey(name: 'person_type') final  PersonType personType;
@override final  RfcInput rfc;
/// Trimmed and NFC-normalized; no invisible characters (control, format such as zero-width or bidi controls, line or paragraph separators; a zero-width joiner inside an emoji sequence is allowed): `400 validation_failed`, field code `invisible_characters`.
@override@JsonKey(name: 'display_name') final  String displayName;
/// At least one regime; each code at most once (`regimes.N.regime_code` / `duplicate`).
 final  List<TaxpayerRegimeCreate> _regimes;
/// At least one regime; each code at most once (`regimes.N.regime_code` / `duplicate`).
@override List<TaxpayerRegimeCreate> get regimes {
  if (_regimes is EqualUnmodifiableListView) return _regimes;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_regimes);
}

/// Optional client-generated id (idempotent retries).
@override@JsonKey(includeIfNull: false) final  Uuid? id;

/// Create a copy of TaxpayerCreate
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TaxpayerCreateCopyWith<_TaxpayerCreate> get copyWith => __$TaxpayerCreateCopyWithImpl<_TaxpayerCreate>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TaxpayerCreateToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TaxpayerCreate&&(identical(other.personType, personType) || other.personType == personType)&&(identical(other.rfc, rfc) || other.rfc == rfc)&&(identical(other.displayName, displayName) || other.displayName == displayName)&&const DeepCollectionEquality().equals(other.regimes, _regimes)&&(identical(other.id, id) || other.id == id));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,personType,rfc,displayName,const DeepCollectionEquality().hash(_regimes),id);
}

@override
String toString() {
    return 'TaxpayerCreate(personType: $personType, rfc: $rfc, displayName: $displayName, regimes: $regimes, id: $id)';
}


}

/// @nodoc
abstract mixin class _$TaxpayerCreateCopyWith<$Res> implements $TaxpayerCreateCopyWith<$Res> {
  factory _$TaxpayerCreateCopyWith(_TaxpayerCreate value, $Res Function(_TaxpayerCreate) _then) = __$TaxpayerCreateCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'person_type') PersonType personType, RfcInput rfc,@JsonKey(name: 'display_name') String displayName, List<TaxpayerRegimeCreate> regimes,@JsonKey(includeIfNull: false) Uuid? id
});




}
/// @nodoc
class __$TaxpayerCreateCopyWithImpl<$Res>
    implements _$TaxpayerCreateCopyWith<$Res> {
  __$TaxpayerCreateCopyWithImpl(this._self, this._then);

  final _TaxpayerCreate _self;
  final $Res Function(_TaxpayerCreate) _then;

/// Create a copy of TaxpayerCreate
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? personType = null,Object? rfc = null,Object? displayName = null,Object? regimes = null,Object? id = freezed,}) {
  return _then(_TaxpayerCreate(
personType: null == personType ? _self.personType : personType // ignore: cast_nullable_to_non_nullable
as PersonType,rfc: null == rfc ? _self.rfc : rfc // ignore: cast_nullable_to_non_nullable
as RfcInput,displayName: null == displayName ? _self.displayName : displayName // ignore: cast_nullable_to_non_nullable
as String,regimes: null == regimes ? _self._regimes : regimes // ignore: cast_nullable_to_non_nullable
as List<TaxpayerRegimeCreate>,id: freezed == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid?,
  ));
}


}

// dart format on
