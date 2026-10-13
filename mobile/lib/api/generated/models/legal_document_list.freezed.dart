// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'legal_document_list.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$LegalDocumentList {

 List<LegalDocument> get items;
/// Create a copy of LegalDocumentList
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$LegalDocumentListCopyWith<LegalDocumentList> get copyWith => _$LegalDocumentListCopyWithImpl<LegalDocumentList>(this as LegalDocumentList, _$identity);

  /// Serializes this LegalDocumentList to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as LegalDocumentList;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is LegalDocumentList&&const DeepCollectionEquality().equals(other.items, _this.items));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as LegalDocumentList;
  return Object.hash(runtimeType,const DeepCollectionEquality().hash(_this.items));
}

@override
String toString() {
  final _this = this as LegalDocumentList;
  return 'LegalDocumentList(items: ${_this.items})';
}


}

/// @nodoc
abstract mixin class $LegalDocumentListCopyWith<$Res>  {
  factory $LegalDocumentListCopyWith(LegalDocumentList value, $Res Function(LegalDocumentList) _then) = _$LegalDocumentListCopyWithImpl;
@useResult
$Res call({
 List<LegalDocument> items
});




}
/// @nodoc
class _$LegalDocumentListCopyWithImpl<$Res>
    implements $LegalDocumentListCopyWith<$Res> {
  _$LegalDocumentListCopyWithImpl(this._self, this._then);

  final LegalDocumentList _self;
  final $Res Function(LegalDocumentList) _then;

/// Create a copy of LegalDocumentList
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? items = null,}) {
  return _then(LegalDocumentList(
items: null == items ? _self.items : items // ignore: cast_nullable_to_non_nullable
as List<LegalDocument>,
  ));
}

}


/// Adds pattern-matching-related methods to [LegalDocumentList].
extension LegalDocumentListPatterns on LegalDocumentList {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _LegalDocumentList value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _LegalDocumentList() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _LegalDocumentList value)  $default,){
final _that = this;
switch (_that) {
case _LegalDocumentList():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _LegalDocumentList value)?  $default,){
final _that = this;
switch (_that) {
case _LegalDocumentList() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( List<LegalDocument> items)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _LegalDocumentList() when $default != null:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( List<LegalDocument> items)  $default,) {final _that = this;
switch (_that) {
case _LegalDocumentList():
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( List<LegalDocument> items)?  $default,) {final _that = this;
switch (_that) {
case _LegalDocumentList() when $default != null:
return $default(_that.items);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _LegalDocumentList implements LegalDocumentList {
  const _LegalDocumentList({required  List<LegalDocument> items}): _items = items;
  factory _LegalDocumentList.fromJson(Map<String, dynamic> json) => _$LegalDocumentListFromJson(json);

 final  List<LegalDocument> _items;
@override List<LegalDocument> get items {
  if (_items is EqualUnmodifiableListView) return _items;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_items);
}


/// Create a copy of LegalDocumentList
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$LegalDocumentListCopyWith<_LegalDocumentList> get copyWith => __$LegalDocumentListCopyWithImpl<_LegalDocumentList>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$LegalDocumentListToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _LegalDocumentList&&const DeepCollectionEquality().equals(other.items, _items));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,const DeepCollectionEquality().hash(_items));
}

@override
String toString() {
    return 'LegalDocumentList(items: $items)';
}


}

/// @nodoc
abstract mixin class _$LegalDocumentListCopyWith<$Res> implements $LegalDocumentListCopyWith<$Res> {
  factory _$LegalDocumentListCopyWith(_LegalDocumentList value, $Res Function(_LegalDocumentList) _then) = __$LegalDocumentListCopyWithImpl;
@override @useResult
$Res call({
 List<LegalDocument> items
});




}
/// @nodoc
class __$LegalDocumentListCopyWithImpl<$Res>
    implements _$LegalDocumentListCopyWith<$Res> {
  __$LegalDocumentListCopyWithImpl(this._self, this._then);

  final _LegalDocumentList _self;
  final $Res Function(_LegalDocumentList) _then;

/// Create a copy of LegalDocumentList
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? items = null,}) {
  return _then(_LegalDocumentList(
items: null == items ? _self._items : items // ignore: cast_nullable_to_non_nullable
as List<LegalDocument>,
  ));
}


}

// dart format on
