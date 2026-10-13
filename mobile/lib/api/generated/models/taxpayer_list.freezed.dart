// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'taxpayer_list.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$TaxpayerList {

 List<Taxpayer> get items;
/// Create a copy of TaxpayerList
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$TaxpayerListCopyWith<TaxpayerList> get copyWith => _$TaxpayerListCopyWithImpl<TaxpayerList>(this as TaxpayerList, _$identity);

  /// Serializes this TaxpayerList to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as TaxpayerList;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is TaxpayerList&&const DeepCollectionEquality().equals(other.items, _this.items));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as TaxpayerList;
  return Object.hash(runtimeType,const DeepCollectionEquality().hash(_this.items));
}

@override
String toString() {
  final _this = this as TaxpayerList;
  return 'TaxpayerList(items: ${_this.items})';
}


}

/// @nodoc
abstract mixin class $TaxpayerListCopyWith<$Res>  {
  factory $TaxpayerListCopyWith(TaxpayerList value, $Res Function(TaxpayerList) _then) = _$TaxpayerListCopyWithImpl;
@useResult
$Res call({
 List<Taxpayer> items
});




}
/// @nodoc
class _$TaxpayerListCopyWithImpl<$Res>
    implements $TaxpayerListCopyWith<$Res> {
  _$TaxpayerListCopyWithImpl(this._self, this._then);

  final TaxpayerList _self;
  final $Res Function(TaxpayerList) _then;

/// Create a copy of TaxpayerList
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? items = null,}) {
  return _then(TaxpayerList(
items: null == items ? _self.items : items // ignore: cast_nullable_to_non_nullable
as List<Taxpayer>,
  ));
}

}


/// Adds pattern-matching-related methods to [TaxpayerList].
extension TaxpayerListPatterns on TaxpayerList {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _TaxpayerList value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _TaxpayerList() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _TaxpayerList value)  $default,){
final _that = this;
switch (_that) {
case _TaxpayerList():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _TaxpayerList value)?  $default,){
final _that = this;
switch (_that) {
case _TaxpayerList() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( List<Taxpayer> items)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _TaxpayerList() when $default != null:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( List<Taxpayer> items)  $default,) {final _that = this;
switch (_that) {
case _TaxpayerList():
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( List<Taxpayer> items)?  $default,) {final _that = this;
switch (_that) {
case _TaxpayerList() when $default != null:
return $default(_that.items);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _TaxpayerList implements TaxpayerList {
  const _TaxpayerList({required  List<Taxpayer> items}): _items = items;
  factory _TaxpayerList.fromJson(Map<String, dynamic> json) => _$TaxpayerListFromJson(json);

 final  List<Taxpayer> _items;
@override List<Taxpayer> get items {
  if (_items is EqualUnmodifiableListView) return _items;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_items);
}


/// Create a copy of TaxpayerList
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$TaxpayerListCopyWith<_TaxpayerList> get copyWith => __$TaxpayerListCopyWithImpl<_TaxpayerList>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$TaxpayerListToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _TaxpayerList&&const DeepCollectionEquality().equals(other.items, _items));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,const DeepCollectionEquality().hash(_items));
}

@override
String toString() {
    return 'TaxpayerList(items: $items)';
}


}

/// @nodoc
abstract mixin class _$TaxpayerListCopyWith<$Res> implements $TaxpayerListCopyWith<$Res> {
  factory _$TaxpayerListCopyWith(_TaxpayerList value, $Res Function(_TaxpayerList) _then) = __$TaxpayerListCopyWithImpl;
@override @useResult
$Res call({
 List<Taxpayer> items
});




}
/// @nodoc
class __$TaxpayerListCopyWithImpl<$Res>
    implements _$TaxpayerListCopyWith<$Res> {
  __$TaxpayerListCopyWithImpl(this._self, this._then);

  final _TaxpayerList _self;
  final $Res Function(_TaxpayerList) _then;

/// Create a copy of TaxpayerList
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? items = null,}) {
  return _then(_TaxpayerList(
items: null == items ? _self._items : items // ignore: cast_nullable_to_non_nullable
as List<Taxpayer>,
  ));
}


}

// dart format on
