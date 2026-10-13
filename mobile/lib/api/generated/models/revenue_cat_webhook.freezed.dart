// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'revenue_cat_webhook.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$RevenueCatWebhook {

@JsonKey(name: 'api_version') String get apiVersion; Event get event;
/// Create a copy of RevenueCatWebhook
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$RevenueCatWebhookCopyWith<RevenueCatWebhook> get copyWith => _$RevenueCatWebhookCopyWithImpl<RevenueCatWebhook>(this as RevenueCatWebhook, _$identity);

  /// Serializes this RevenueCatWebhook to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as RevenueCatWebhook;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is RevenueCatWebhook&&(identical(other.apiVersion, _this.apiVersion) || other.apiVersion == _this.apiVersion)&&(identical(other.event, _this.event) || other.event == _this.event));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as RevenueCatWebhook;
  return Object.hash(runtimeType,_this.apiVersion,_this.event);
}

@override
String toString() {
  final _this = this as RevenueCatWebhook;
  return 'RevenueCatWebhook(apiVersion: ${_this.apiVersion}, event: ${_this.event})';
}


}

/// @nodoc
abstract mixin class $RevenueCatWebhookCopyWith<$Res>  {
  factory $RevenueCatWebhookCopyWith(RevenueCatWebhook value, $Res Function(RevenueCatWebhook) _then) = _$RevenueCatWebhookCopyWithImpl;
@useResult
$Res call({
@JsonKey(name: 'api_version') String apiVersion, Event event
});


$EventCopyWith<$Res> get event;

}
/// @nodoc
class _$RevenueCatWebhookCopyWithImpl<$Res>
    implements $RevenueCatWebhookCopyWith<$Res> {
  _$RevenueCatWebhookCopyWithImpl(this._self, this._then);

  final RevenueCatWebhook _self;
  final $Res Function(RevenueCatWebhook) _then;

/// Create a copy of RevenueCatWebhook
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? apiVersion = null,Object? event = null,}) {
  return _then(RevenueCatWebhook(
apiVersion: null == apiVersion ? _self.apiVersion : apiVersion // ignore: cast_nullable_to_non_nullable
as String,event: null == event ? _self.event : event // ignore: cast_nullable_to_non_nullable
as Event,
  ));
}
/// Create a copy of RevenueCatWebhook
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$EventCopyWith<$Res> get event {
  
  return $EventCopyWith<$Res>(_self.event, (value) {
    return _then(_self.copyWith(event: value));
  });
}
}


/// Adds pattern-matching-related methods to [RevenueCatWebhook].
extension RevenueCatWebhookPatterns on RevenueCatWebhook {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _RevenueCatWebhook value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _RevenueCatWebhook() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _RevenueCatWebhook value)  $default,){
final _that = this;
switch (_that) {
case _RevenueCatWebhook():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _RevenueCatWebhook value)?  $default,){
final _that = this;
switch (_that) {
case _RevenueCatWebhook() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function(@JsonKey(name: 'api_version')  String apiVersion,  Event event)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _RevenueCatWebhook() when $default != null:
return $default(_that.apiVersion,_that.event);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function(@JsonKey(name: 'api_version')  String apiVersion,  Event event)  $default,) {final _that = this;
switch (_that) {
case _RevenueCatWebhook():
return $default(_that.apiVersion,_that.event);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function(@JsonKey(name: 'api_version')  String apiVersion,  Event event)?  $default,) {final _that = this;
switch (_that) {
case _RevenueCatWebhook() when $default != null:
return $default(_that.apiVersion,_that.event);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _RevenueCatWebhook implements RevenueCatWebhook {
  const _RevenueCatWebhook({@JsonKey(name: 'api_version') required this.apiVersion, required this.event});
  factory _RevenueCatWebhook.fromJson(Map<String, dynamic> json) => _$RevenueCatWebhookFromJson(json);

@override@JsonKey(name: 'api_version') final  String apiVersion;
@override final  Event event;

/// Create a copy of RevenueCatWebhook
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$RevenueCatWebhookCopyWith<_RevenueCatWebhook> get copyWith => __$RevenueCatWebhookCopyWithImpl<_RevenueCatWebhook>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$RevenueCatWebhookToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _RevenueCatWebhook&&(identical(other.apiVersion, apiVersion) || other.apiVersion == apiVersion)&&(identical(other.event, event) || other.event == event));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,apiVersion,event);
}

@override
String toString() {
    return 'RevenueCatWebhook(apiVersion: $apiVersion, event: $event)';
}


}

/// @nodoc
abstract mixin class _$RevenueCatWebhookCopyWith<$Res> implements $RevenueCatWebhookCopyWith<$Res> {
  factory _$RevenueCatWebhookCopyWith(_RevenueCatWebhook value, $Res Function(_RevenueCatWebhook) _then) = __$RevenueCatWebhookCopyWithImpl;
@override @useResult
$Res call({
@JsonKey(name: 'api_version') String apiVersion, Event event
});


@override $EventCopyWith<$Res> get event;

}
/// @nodoc
class __$RevenueCatWebhookCopyWithImpl<$Res>
    implements _$RevenueCatWebhookCopyWith<$Res> {
  __$RevenueCatWebhookCopyWithImpl(this._self, this._then);

  final _RevenueCatWebhook _self;
  final $Res Function(_RevenueCatWebhook) _then;

/// Create a copy of RevenueCatWebhook
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? apiVersion = null,Object? event = null,}) {
  return _then(_RevenueCatWebhook(
apiVersion: null == apiVersion ? _self.apiVersion : apiVersion // ignore: cast_nullable_to_non_nullable
as String,event: null == event ? _self.event : event // ignore: cast_nullable_to_non_nullable
as Event,
  ));
}

/// Create a copy of RevenueCatWebhook
/// with the given fields replaced by the non-null parameter values.
@override
@pragma('vm:prefer-inline')
$EventCopyWith<$Res> get event {
  
  return $EventCopyWith<$Res>(_self.event, (value) {
    return _then(_self.copyWith(event: value));
  });
}
}

// dart format on
