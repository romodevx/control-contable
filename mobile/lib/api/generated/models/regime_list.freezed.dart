// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'regime_list.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$RegimeList {

 List<Regime> get items;
/// Create a copy of RegimeList
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$RegimeListCopyWith<RegimeList> get copyWith => _$RegimeListCopyWithImpl<RegimeList>(this as RegimeList, _$identity);

  /// Serializes this RegimeList to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as RegimeList;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is RegimeList&&const DeepCollectionEquality().equals(other.items, _this.items));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as RegimeList;
  return Object.hash(runtimeType,const DeepCollectionEquality().hash(_this.items));
}

@override
String toString() {
  final _this = this as RegimeList;
  return 'RegimeList(items: ${_this.items})';
}


}

/// @nodoc
abstract mixin class $RegimeListCopyWith<$Res>  {
  factory $RegimeListCopyWith(RegimeList value, $Res Function(RegimeList) _then) = _$RegimeListCopyWithImpl;
@useResult
$Res call({
 List<Regime> items
});




}
/// @nodoc
class _$RegimeListCopyWithImpl<$Res>
    implements $RegimeListCopyWith<$Res> {
  _$RegimeListCopyWithImpl(this._self, this._then);

  final RegimeList _self;
  final $Res Function(RegimeList) _then;

/// Create a copy of RegimeList
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? items = null,}) {
  return _then(RegimeList(
items: null == items ? _self.items : items // ignore: cast_nullable_to_non_nullable
as List<Regime>,
  ));
}

}


/// Adds pattern-matching-related methods to [RegimeList].
extension RegimeListPatterns on RegimeList {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _RegimeList value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _RegimeList() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _RegimeList value)  $default,){
final _that = this;
switch (_that) {
case _RegimeList():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _RegimeList value)?  $default,){
final _that = this;
switch (_that) {
case _RegimeList() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( List<Regime> items)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _RegimeList() when $default != null:
return $default(_that.items);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( List<Regime> items)  $default,) {final _that = this;
switch (_that) {
case _RegimeList():
return $default(_that.items);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( List<Regime> items)?  $default,) {final _that = this;
switch (_that) {
case _RegimeList() when $default != null:
return $default(_that.items);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _RegimeList implements RegimeList {
  const _RegimeList({required  List<Regime> items}): _items = items;
  factory _RegimeList.fromJson(Map<String, dynamic> json) => _$RegimeListFromJson(json);

 final  List<Regime> _items;
@override List<Regime> get items {
  if (_items is EqualUnmodifiableListView) return _items;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_items);
}


/// Create a copy of RegimeList
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$RegimeListCopyWith<_RegimeList> get copyWith => __$RegimeListCopyWithImpl<_RegimeList>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$RegimeListToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _RegimeList&&const DeepCollectionEquality().equals(other.items, _items));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,const DeepCollectionEquality().hash(_items));
}

@override
String toString() {
    return 'RegimeList(items: $items)';
}


}

/// @nodoc
abstract mixin class _$RegimeListCopyWith<$Res> implements $RegimeListCopyWith<$Res> {
  factory _$RegimeListCopyWith(_RegimeList value, $Res Function(_RegimeList) _then) = __$RegimeListCopyWithImpl;
@override @useResult
$Res call({
 List<Regime> items
});




}
/// @nodoc
class __$RegimeListCopyWithImpl<$Res>
    implements _$RegimeListCopyWith<$Res> {
  __$RegimeListCopyWithImpl(this._self, this._then);

  final _RegimeList _self;
  final $Res Function(_RegimeList) _then;

/// Create a copy of RegimeList
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? items = null,}) {
  return _then(_RegimeList(
items: null == items ? _self._items : items // ignore: cast_nullable_to_non_nullable
as List<Regime>,
  ));
}


}

// dart format on
