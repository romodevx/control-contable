// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'expense_category_list.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$ExpenseCategoryList {

 List<ExpenseCategory> get items; Legend get legend;
/// Create a copy of ExpenseCategoryList
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ExpenseCategoryListCopyWith<ExpenseCategoryList> get copyWith => _$ExpenseCategoryListCopyWithImpl<ExpenseCategoryList>(this as ExpenseCategoryList, _$identity);

  /// Serializes this ExpenseCategoryList to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as ExpenseCategoryList;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is ExpenseCategoryList&&const DeepCollectionEquality().equals(other.items, _this.items)&&(identical(other.legend, _this.legend) || other.legend == _this.legend));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as ExpenseCategoryList;
  return Object.hash(runtimeType,const DeepCollectionEquality().hash(_this.items),_this.legend);
}

@override
String toString() {
  final _this = this as ExpenseCategoryList;
  return 'ExpenseCategoryList(items: ${_this.items}, legend: ${_this.legend})';
}


}

/// @nodoc
abstract mixin class $ExpenseCategoryListCopyWith<$Res>  {
  factory $ExpenseCategoryListCopyWith(ExpenseCategoryList value, $Res Function(ExpenseCategoryList) _then) = _$ExpenseCategoryListCopyWithImpl;
@useResult
$Res call({
 List<ExpenseCategory> items, Legend legend
});




}
/// @nodoc
class _$ExpenseCategoryListCopyWithImpl<$Res>
    implements $ExpenseCategoryListCopyWith<$Res> {
  _$ExpenseCategoryListCopyWithImpl(this._self, this._then);

  final ExpenseCategoryList _self;
  final $Res Function(ExpenseCategoryList) _then;

/// Create a copy of ExpenseCategoryList
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? items = null,Object? legend = null,}) {
  return _then(ExpenseCategoryList(
items: null == items ? _self.items : items // ignore: cast_nullable_to_non_nullable
as List<ExpenseCategory>,legend: null == legend ? _self.legend : legend // ignore: cast_nullable_to_non_nullable
as Legend,
  ));
}

}


/// Adds pattern-matching-related methods to [ExpenseCategoryList].
extension ExpenseCategoryListPatterns on ExpenseCategoryList {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _ExpenseCategoryList value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _ExpenseCategoryList() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _ExpenseCategoryList value)  $default,){
final _that = this;
switch (_that) {
case _ExpenseCategoryList():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _ExpenseCategoryList value)?  $default,){
final _that = this;
switch (_that) {
case _ExpenseCategoryList() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( List<ExpenseCategory> items,  Legend legend)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _ExpenseCategoryList() when $default != null:
return $default(_that.items,_that.legend);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( List<ExpenseCategory> items,  Legend legend)  $default,) {final _that = this;
switch (_that) {
case _ExpenseCategoryList():
return $default(_that.items,_that.legend);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( List<ExpenseCategory> items,  Legend legend)?  $default,) {final _that = this;
switch (_that) {
case _ExpenseCategoryList() when $default != null:
return $default(_that.items,_that.legend);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _ExpenseCategoryList implements ExpenseCategoryList {
  const _ExpenseCategoryList({required  List<ExpenseCategory> items, required this.legend}): _items = items;
  factory _ExpenseCategoryList.fromJson(Map<String, dynamic> json) => _$ExpenseCategoryListFromJson(json);

 final  List<ExpenseCategory> _items;
@override List<ExpenseCategory> get items {
  if (_items is EqualUnmodifiableListView) return _items;
  // ignore: implicit_dynamic_type
  return EqualUnmodifiableListView(_items);
}

@override final  Legend legend;

/// Create a copy of ExpenseCategoryList
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ExpenseCategoryListCopyWith<_ExpenseCategoryList> get copyWith => __$ExpenseCategoryListCopyWithImpl<_ExpenseCategoryList>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$ExpenseCategoryListToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _ExpenseCategoryList&&const DeepCollectionEquality().equals(other.items, _items)&&(identical(other.legend, legend) || other.legend == legend));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,const DeepCollectionEquality().hash(_items),legend);
}

@override
String toString() {
    return 'ExpenseCategoryList(items: $items, legend: $legend)';
}


}

/// @nodoc
abstract mixin class _$ExpenseCategoryListCopyWith<$Res> implements $ExpenseCategoryListCopyWith<$Res> {
  factory _$ExpenseCategoryListCopyWith(_ExpenseCategoryList value, $Res Function(_ExpenseCategoryList) _then) = __$ExpenseCategoryListCopyWithImpl;
@override @useResult
$Res call({
 List<ExpenseCategory> items, Legend legend
});




}
/// @nodoc
class __$ExpenseCategoryListCopyWithImpl<$Res>
    implements _$ExpenseCategoryListCopyWith<$Res> {
  __$ExpenseCategoryListCopyWithImpl(this._self, this._then);

  final _ExpenseCategoryList _self;
  final $Res Function(_ExpenseCategoryList) _then;

/// Create a copy of ExpenseCategoryList
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? items = null,Object? legend = null,}) {
  return _then(_ExpenseCategoryList(
items: null == items ? _self._items : items // ignore: cast_nullable_to_non_nullable
as List<ExpenseCategory>,legend: null == legend ? _self.legend : legend // ignore: cast_nullable_to_non_nullable
as Legend,
  ));
}


}

// dart format on
