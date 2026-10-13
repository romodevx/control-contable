// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'taxpayer.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$Taxpayer {

 Uuid get id;@JsonKey(name: 'person_type') PersonType get personType;@JsonKey(name: 'display_name') String get displayName;/// The full RFC is never returned.
@JsonKey(name: 'rfc_masked') String get rfcMasked; List<TaxpayerRegime> get regimes;@JsonKey(name: 'created_at') DateTime get createdAt;@JsonKey(name: 'archived_at') DateTime? get archivedAt;
/// Create a copy of Taxpayer
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TaxpayerCopyWith<Taxpayer> get copyWith => _$TaxpayerCopyWithImpl<Taxpayer>(this as Taxpayer, _$identity);

  /// Serializes this Taxpayer to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as Taxpayer;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is Taxpayer&&(identical(other.id, _this.id) || other.id == _this.id)&&(identical(other.personType, _this.personType) || other.personType == _this.personType)&&(identical(other.displayName, _this.displayName) || other.displayName == _this.displayName)&&(identical(other.rfcMasked, _this.rfcMasked) || other.rfcMasked == _this.rfcMasked)&&const DeepCollectionEquality().equals(other.regimes, _this.regimes)&&(identical(other.createdAt, _this.createdAt) || other.createdAt == _this.createdAt)&&(identical(other.archivedAt, _this.archivedAt) || other.archivedAt == _this.archivedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as Taxpayer;
  return Object.hash(runtimeType,_this.id,_this.personType,_this.displayName,_this.rfcMasked,const DeepCollectionEquality().hash(_this.regimes),_this.createdAt,_this.archivedAt);
}

@override
String toString() {
  final _this = this as Taxpayer;
  return 'Taxpayer(id: ${_this.id}, personType: ${_this.personType}, displayName: ${_this.displayName}, rfcMasked: ${_this.rfcMasked}, regimes: ${_this.regimes}, createdAt: ${_this.createdAt}, archivedAt: ${_this.archivedAt})';
}


}

/// @nodoc
abstract mixin class $TaxpayerCopyWith<$Res>  {
  factory $TaxpayerCopyWith(Taxpayer value, $Res Function(Taxpayer) _then) = _$TaxpayerCopyWithImpl;
@useResult
$Res call({
 Uuid id,@JsonKey(name: 'person_type') PersonType personType,@JsonKey(name: 'display_name') String displayName,@JsonKey(name: 'rfc_masked') String rfcMasked, List<TaxpayerRegime> regimes,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'archived_at') DateTime? archivedAt
});




}
/// @nodoc
class _$TaxpayerCopyWithImpl<$Res>
    implements $TaxpayerCopyWith<$Res> {
  _$TaxpayerCopyWithImpl(this._self, this._then);

  final Taxpayer _self;
  final $Res Function(Taxpayer) _then;

/// Create a copy of Taxpayer
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? id = null,Object? personType = null,Object? displayName = null,Object? rfcMasked = null,Object? regimes = null,Object? createdAt = null,Object? archivedAt = freezed,}) {
  return _then(Taxpayer(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid,personType: null == personType ? _self.personType : personType // ignore: cast_nullable_to_non_nullable
as PersonType,displayName: null == displayName ? _self.displayName : displayName // ignore: cast_nullable_to_non_nullable
as String,rfcMasked: null == rfcMasked ? _self.rfcMasked : rfcMasked // ignore: cast_nullable_to_non_nullable
as String,regimes: null == regimes ? _self.regimes : regimes // ignore: cast_nullable_to_non_nullable
as List<TaxpayerRegime>,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,archivedAt: freezed == archivedAt ? _self.archivedAt : archivedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}

}


/// Adds pattern-matching-related methods to [Taxpayer].
extension TaxpayerPatterns on Taxpayer {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _Taxpayer value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _Taxpayer() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _Taxpayer value)  $default,){
final _that = this;
switch (_that) {
case _Taxpayer():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _Taxpayer value)?  $default,){
final _that = this;
switch (_that) {
case _Taxpayer() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( Uuid id, @JsonKey(name: 'person_type')  PersonType personType, @JsonKey(name: 'display_name')  String displayName, @JsonKey(name: 'rfc_masked')  String rfcMasked,  List<TaxpayerRegime> regimes, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'archived_at')  DateTime? archivedAt)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _Taxpayer() when $default != null:
return $default(_that.id,_that.personType,_that.displayName,_that.rfcMasked,_that.regimes,_that.createdAt,_that.archivedAt);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( Uuid id, @JsonKey(name: 'person_type')  PersonType personType, @JsonKey(name: 'display_name')  String displayName, @JsonKey(name: 'rfc_masked')  String rfcMasked,  List<TaxpayerRegime> regimes, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'archived_at')  DateTime? archivedAt)  $default,) {final _that = this;
switch (_that) {
case _Taxpayer():
return $default(_that.id,_that.personType,_that.displayName,_that.rfcMasked,_that.regimes,_that.createdAt,_that.archivedAt);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( Uuid id, @JsonKey(name: 'person_type')  PersonType personType, @JsonKey(name: 'display_name')  String displayName, @JsonKey(name: 'rfc_masked')  String rfcMasked,  List<TaxpayerRegime> regimes, @JsonKey(name: 'created_at')  DateTime createdAt, @JsonKey(name: 'archived_at')  DateTime? archivedAt)?  $default,) {final _that = this;
switch (_that) {
case _Taxpayer() when $default != null:
return $default(_that.id,_that.personType,_that.displayName,_that.rfcMasked,_that.regimes,_that.createdAt,_that.archivedAt);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _Taxpayer implements Taxpayer {
  const _Taxpayer({required this.id, @JsonKey(name: 'person_type') required this.personType, @JsonKey(name: 'display_name') required this.displayName, @JsonKey(name: 'rfc_masked') required this.rfcMasked, required  List<TaxpayerRegime> regimes, @JsonKey(name: 'created_at') required this.createdAt, @JsonKey(name: 'archived_at') this.archivedAt}): _regimes = regimes;
  factory _Taxpayer.fromJson(Map<String, dynamic> json) => _$TaxpayerFromJson(json);

@override final  Uuid id;
@override@JsonKey(name: 'person_type') final  PersonType personType;
@override@JsonKey(name: 'display_name') final  String displayName;
/// The full RFC is never returned.
@override@JsonKey(name: 'rfc_masked') final  String rfcMasked;
 final  List<TaxpayerRegime> _regimes;
@override List<TaxpayerRegime> get regimes {
  if (_regimes is EqualUnmodifiableListView) return _regimes;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_regimes);
}

@override@JsonKey(name: 'created_at') final  DateTime createdAt;
@override@JsonKey(name: 'archived_at') final  DateTime? archivedAt;

/// Create a copy of Taxpayer
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TaxpayerCopyWith<_Taxpayer> get copyWith => __$TaxpayerCopyWithImpl<_Taxpayer>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TaxpayerToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _Taxpayer&&(identical(other.id, id) || other.id == id)&&(identical(other.personType, personType) || other.personType == personType)&&(identical(other.displayName, displayName) || other.displayName == displayName)&&(identical(other.rfcMasked, rfcMasked) || other.rfcMasked == rfcMasked)&&const DeepCollectionEquality().equals(other.regimes, _regimes)&&(identical(other.createdAt, createdAt) || other.createdAt == createdAt)&&(identical(other.archivedAt, archivedAt) || other.archivedAt == archivedAt));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,id,personType,displayName,rfcMasked,const DeepCollectionEquality().hash(_regimes),createdAt,archivedAt);
}

@override
String toString() {
    return 'Taxpayer(id: $id, personType: $personType, displayName: $displayName, rfcMasked: $rfcMasked, regimes: $regimes, createdAt: $createdAt, archivedAt: $archivedAt)';
}


}

/// @nodoc
abstract mixin class _$TaxpayerCopyWith<$Res> implements $TaxpayerCopyWith<$Res> {
  factory _$TaxpayerCopyWith(_Taxpayer value, $Res Function(_Taxpayer) _then) = __$TaxpayerCopyWithImpl;
@override @useResult
$Res call({
 Uuid id,@JsonKey(name: 'person_type') PersonType personType,@JsonKey(name: 'display_name') String displayName,@JsonKey(name: 'rfc_masked') String rfcMasked, List<TaxpayerRegime> regimes,@JsonKey(name: 'created_at') DateTime createdAt,@JsonKey(name: 'archived_at') DateTime? archivedAt
});




}
/// @nodoc
class __$TaxpayerCopyWithImpl<$Res>
    implements _$TaxpayerCopyWith<$Res> {
  __$TaxpayerCopyWithImpl(this._self, this._then);

  final _Taxpayer _self;
  final $Res Function(_Taxpayer) _then;

/// Create a copy of Taxpayer
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? id = null,Object? personType = null,Object? displayName = null,Object? rfcMasked = null,Object? regimes = null,Object? createdAt = null,Object? archivedAt = freezed,}) {
  return _then(_Taxpayer(
id: null == id ? _self.id : id // ignore: cast_nullable_to_non_nullable
as Uuid,personType: null == personType ? _self.personType : personType // ignore: cast_nullable_to_non_nullable
as PersonType,displayName: null == displayName ? _self.displayName : displayName // ignore: cast_nullable_to_non_nullable
as String,rfcMasked: null == rfcMasked ? _self.rfcMasked : rfcMasked // ignore: cast_nullable_to_non_nullable
as String,regimes: null == regimes ? _self._regimes : regimes // ignore: cast_nullable_to_non_nullable
as List<TaxpayerRegime>,createdAt: null == createdAt ? _self.createdAt : createdAt // ignore: cast_nullable_to_non_nullable
as DateTime,archivedAt: freezed == archivedAt ? _self.archivedAt : archivedAt // ignore: cast_nullable_to_non_nullable
as DateTime?,
  ));
}


}

// dart format on
