// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'token_pair.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$TokenPair {

@JsonKey(name: 'access_token') String get accessToken;/// Seconds.
@JsonKey(name: 'access_expires_in') int get accessExpiresIn;/// Opaque, format <user_uuid>.<secret>.
@JsonKey(name: 'refresh_token') String get refreshToken;/// Seconds.
@JsonKey(name: 'refresh_expires_in') int get refreshExpiresIn;@JsonKey(name: 'token_type') String get tokenType;
/// Create a copy of TokenPair
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TokenPairCopyWith<TokenPair> get copyWith => _$TokenPairCopyWithImpl<TokenPair>(this as TokenPair, _$identity);

  /// Serializes this TokenPair to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as TokenPair;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TokenPair&&(identical(other.accessToken, _this.accessToken) || other.accessToken == _this.accessToken)&&(identical(other.accessExpiresIn, _this.accessExpiresIn) || other.accessExpiresIn == _this.accessExpiresIn)&&(identical(other.refreshToken, _this.refreshToken) || other.refreshToken == _this.refreshToken)&&(identical(other.refreshExpiresIn, _this.refreshExpiresIn) || other.refreshExpiresIn == _this.refreshExpiresIn)&&(identical(other.tokenType, _this.tokenType) || other.tokenType == _this.tokenType));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TokenPair;
  return Object.hash(runtimeType,_this.accessToken,_this.accessExpiresIn,_this.refreshToken,_this.refreshExpiresIn,_this.tokenType);
}

@override
String toString() {
  final _this = this as TokenPair;
  return 'TokenPair(accessToken: ${_this.accessToken}, accessExpiresIn: ${_this.accessExpiresIn}, refreshToken: ${_this.refreshToken}, refreshExpiresIn: ${_this.refreshExpiresIn}, tokenType: ${_this.tokenType})';
}


}

/// @nodoc
abstract mixin class $TokenPairCopyWith<$Res>  {
  factory $TokenPairCopyWith(TokenPair value, $Res Function(TokenPair) _then) = _$TokenPairCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'access_token') String accessToken,@JsonKey(name: 'access_expires_in') int accessExpiresIn,@JsonKey(name: 'refresh_token') String refreshToken,@JsonKey(name: 'refresh_expires_in') int refreshExpiresIn,@JsonKey(name: 'token_type') String tokenType
});




}
/// @nodoc
class _$TokenPairCopyWithImpl<$Res>
    implements $TokenPairCopyWith<$Res> {
  _$TokenPairCopyWithImpl(this._self, this._then);

  final TokenPair _self;
  final $Res Function(TokenPair) _then;

/// Create a copy of TokenPair
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? accessToken = null,Object? accessExpiresIn = null,Object? refreshToken = null,Object? refreshExpiresIn = null,Object? tokenType = null,}) {
  return _then(TokenPair(
accessToken: null == accessToken ? _self.accessToken : accessToken // ignore: cast_nullable_to_non_nullable
as String,accessExpiresIn: null == accessExpiresIn ? _self.accessExpiresIn : accessExpiresIn // ignore: cast_nullable_to_non_nullable
as int,refreshToken: null == refreshToken ? _self.refreshToken : refreshToken // ignore: cast_nullable_to_non_nullable
as String,refreshExpiresIn: null == refreshExpiresIn ? _self.refreshExpiresIn : refreshExpiresIn // ignore: cast_nullable_to_non_nullable
as int,tokenType: null == tokenType ? _self.tokenType : tokenType // ignore: cast_nullable_to_non_nullable
as String,
  ));
}

}


/// Adds pattern-matching-related methods to [TokenPair].
extension TokenPairPatterns on TokenPair {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _TokenPair value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _TokenPair() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _TokenPair value)  $default,){
final _that = this;
switch (_that) {
case _TokenPair():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _TokenPair value)?  $default,){
final _that = this;
switch (_that) {
case _TokenPair() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'access_token')  String accessToken, @JsonKey(name: 'access_expires_in')  int accessExpiresIn, @JsonKey(name: 'refresh_token')  String refreshToken, @JsonKey(name: 'refresh_expires_in')  int refreshExpiresIn, @JsonKey(name: 'token_type')  String tokenType)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TokenPair() when $default != null:
return $default(_that.accessToken,_that.accessExpiresIn,_that.refreshToken,_that.refreshExpiresIn,_that.tokenType);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'access_token')  String accessToken, @JsonKey(name: 'access_expires_in')  int accessExpiresIn, @JsonKey(name: 'refresh_token')  String refreshToken, @JsonKey(name: 'refresh_expires_in')  int refreshExpiresIn, @JsonKey(name: 'token_type')  String tokenType)  $default,) {final _that = this;
switch (_that) {
case _TokenPair():
return $default(_that.accessToken,_that.accessExpiresIn,_that.refreshToken,_that.refreshExpiresIn,_that.tokenType);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'access_token')  String accessToken, @JsonKey(name: 'access_expires_in')  int accessExpiresIn, @JsonKey(name: 'refresh_token')  String refreshToken, @JsonKey(name: 'refresh_expires_in')  int refreshExpiresIn, @JsonKey(name: 'token_type')  String tokenType)?  $default,) {final _that = this;
switch (_that) {
case _TokenPair() when $default != null:
return $default(_that.accessToken,_that.accessExpiresIn,_that.refreshToken,_that.refreshExpiresIn,_that.tokenType);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TokenPair implements TokenPair {
  const _TokenPair({@JsonKey(name: 'access_token') required this.accessToken, @JsonKey(name: 'access_expires_in') required this.accessExpiresIn, @JsonKey(name: 'refresh_token') required this.refreshToken, @JsonKey(name: 'refresh_expires_in') required this.refreshExpiresIn, @JsonKey(name: 'token_type') required this.tokenType});
  factory _TokenPair.fromJson(Map<String, dynamic> json) => _$TokenPairFromJson(json);

@override@JsonKey(name: 'access_token') final  String accessToken;
/// Seconds.
@override@JsonKey(name: 'access_expires_in') final  int accessExpiresIn;
/// Opaque, format <user_uuid>.<secret>.
@override@JsonKey(name: 'refresh_token') final  String refreshToken;
/// Seconds.
@override@JsonKey(name: 'refresh_expires_in') final  int refreshExpiresIn;
@override@JsonKey(name: 'token_type') final  String tokenType;

/// Create a copy of TokenPair
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TokenPairCopyWith<_TokenPair> get copyWith => __$TokenPairCopyWithImpl<_TokenPair>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TokenPairToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TokenPair&&(identical(other.accessToken, accessToken) || other.accessToken == accessToken)&&(identical(other.accessExpiresIn, accessExpiresIn) || other.accessExpiresIn == accessExpiresIn)&&(identical(other.refreshToken, refreshToken) || other.refreshToken == refreshToken)&&(identical(other.refreshExpiresIn, refreshExpiresIn) || other.refreshExpiresIn == refreshExpiresIn)&&(identical(other.tokenType, tokenType) || other.tokenType == tokenType));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,accessToken,accessExpiresIn,refreshToken,refreshExpiresIn,tokenType);
}

@override
String toString() {
    return 'TokenPair(accessToken: $accessToken, accessExpiresIn: $accessExpiresIn, refreshToken: $refreshToken, refreshExpiresIn: $refreshExpiresIn, tokenType: $tokenType)';
}


}

/// @nodoc
abstract mixin class _$TokenPairCopyWith<$Res> implements $TokenPairCopyWith<$Res> {
  factory _$TokenPairCopyWith(_TokenPair value, $Res Function(_TokenPair) _then) = __$TokenPairCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'access_token') String accessToken,@JsonKey(name: 'access_expires_in') int accessExpiresIn,@JsonKey(name: 'refresh_token') String refreshToken,@JsonKey(name: 'refresh_expires_in') int refreshExpiresIn,@JsonKey(name: 'token_type') String tokenType
});




}
/// @nodoc
class __$TokenPairCopyWithImpl<$Res>
    implements _$TokenPairCopyWith<$Res> {
  __$TokenPairCopyWithImpl(this._self, this._then);

  final _TokenPair _self;
  final $Res Function(_TokenPair) _then;

/// Create a copy of TokenPair
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? accessToken = null,Object? accessExpiresIn = null,Object? refreshToken = null,Object? refreshExpiresIn = null,Object? tokenType = null,}) {
  return _then(_TokenPair(
accessToken: null == accessToken ? _self.accessToken : accessToken // ignore: cast_nullable_to_non_nullable
as String,accessExpiresIn: null == accessExpiresIn ? _self.accessExpiresIn : accessExpiresIn // ignore: cast_nullable_to_non_nullable
as int,refreshToken: null == refreshToken ? _self.refreshToken : refreshToken // ignore: cast_nullable_to_non_nullable
as String,refreshExpiresIn: null == refreshExpiresIn ? _self.refreshExpiresIn : refreshExpiresIn // ignore: cast_nullable_to_non_nullable
as int,tokenType: null == tokenType ? _self.tokenType : tokenType // ignore: cast_nullable_to_non_nullable
as String,
  ));
}


}

// dart format on
