// GENERATED CODE - DO NOT MODIFY BY HAND
// coverage:ignore-file
// ignore_for_file: type=lint, type=warning, deprecated_member_use, deprecated_member_use_from_same_package
// ignore_for_file: unused_element, deprecated_member_use, deprecated_member_use_from_same_package, use_function_type_syntax_for_parameters, unnecessary_const, avoid_init_to_null, invalid_override_different_default_values_named, prefer_expression_function_bodies, annotate_overrides, invalid_annotation_target, unnecessary_question_mark

part of 'expense_category.dart';

// **************************************************************************
// FreezedGenerator
// **************************************************************************

// GENERATED CODE - DO NOT MODIFY BY HAND
// dart format off
T _$identity<T>(T value) => value;

/// @nodoc
mixin _$ExpenseCategory {

 String get code;/// es-MX display name.
 String get name; Deductibility get deductibility;/// The `deductible` a new expense starts with (before the CFDI and cash rules); conservative for partial cases.
@JsonKey(name: 'default_deductible') bool get defaultDeductible;/// es-MX note shown with the hint.
@JsonKey(includeIfNull: true) String? get note;/// Article-level basis (e.g. `LISR 103`).
@JsonKey(includeIfNull: true, name: 'legal_basis') String? get legalBasis;/// The `operation_kind` of this category (seed data in `category_regime_rules`, MODULE-withholdings §2.2): what a persona moral's "Tu proveedor" question E3 starts at, and what `createExpense`, `updateExpense` and `previewExpenseWithholdings` use when `operation_kind` is omitted. Clients never keep their own copy of this table. Optional in 0.10.1 (additive); sent once WH-03 seeds it and WH-07 serves it, when it **will be required**. Absent, a client leaves E3 unselected and omits `operation_kind`, so the server applies the same default.
@JsonKey(includeIfNull: false, name: 'default_operation_kind') OperationKind? get defaultOperationKind;
/// Create a copy of ExpenseCategory
/// with the given fields replaced by the non-null parameter values.
@JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
$ExpenseCategoryCopyWith<ExpenseCategory> get copyWith => _$ExpenseCategoryCopyWithImpl<ExpenseCategory>(this as ExpenseCategory, _$identity);

  /// Serializes this ExpenseCategory to a JSON map.
  Map<String, dynamic> toJson();


@override
bool operator ==(Object other) {
  final _this = this as ExpenseCategory;
  return identical(this, other) || (other.runtimeType == runtimeType&&other is ExpenseCategory&&(identical(other.code, _this.code) || other.code == _this.code)&&(identical(other.name, _this.name) || other.name == _this.name)&&(identical(other.deductibility, _this.deductibility) || other.deductibility == _this.deductibility)&&(identical(other.defaultDeductible, _this.defaultDeductible) || other.defaultDeductible == _this.defaultDeductible)&&(identical(other.note, _this.note) || other.note == _this.note)&&(identical(other.legalBasis, _this.legalBasis) || other.legalBasis == _this.legalBasis)&&(identical(other.defaultOperationKind, _this.defaultOperationKind) || other.defaultOperationKind == _this.defaultOperationKind));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
  final _this = this as ExpenseCategory;
  return Object.hash(runtimeType,_this.code,_this.name,_this.deductibility,_this.defaultDeductible,_this.note,_this.legalBasis,_this.defaultOperationKind);
}

@override
String toString() {
  final _this = this as ExpenseCategory;
  return 'ExpenseCategory(code: ${_this.code}, name: ${_this.name}, deductibility: ${_this.deductibility}, defaultDeductible: ${_this.defaultDeductible}, note: ${_this.note}, legalBasis: ${_this.legalBasis}, defaultOperationKind: ${_this.defaultOperationKind})';
}


}

/// @nodoc
abstract mixin class $ExpenseCategoryCopyWith<$Res>  {
  factory $ExpenseCategoryCopyWith(ExpenseCategory value, $Res Function(ExpenseCategory) _then) = _$ExpenseCategoryCopyWithImpl;
@useResult
$Res call({
 String code, String name, Deductibility deductibility,@JsonKey(name: 'default_deductible') bool defaultDeductible,@JsonKey(includeIfNull: true) String? note,@JsonKey(includeIfNull: true, name: 'legal_basis') String? legalBasis,@JsonKey(includeIfNull: false, name: 'default_operation_kind') OperationKind? defaultOperationKind
});




}
/// @nodoc
class _$ExpenseCategoryCopyWithImpl<$Res>
    implements $ExpenseCategoryCopyWith<$Res> {
  _$ExpenseCategoryCopyWithImpl(this._self, this._then);

  final ExpenseCategory _self;
  final $Res Function(ExpenseCategory) _then;

/// Create a copy of ExpenseCategory
/// with the given fields replaced by the non-null parameter values.
@pragma('vm:prefer-inline') @override $Res call({Object? code = null,Object? name = null,Object? deductibility = null,Object? defaultDeductible = null,Object? note = freezed,Object? legalBasis = freezed,Object? defaultOperationKind = freezed,}) {
  return _then(ExpenseCategory(
code: null == code ? _self.code : code // ignore: cast_nullable_to_non_nullable
as String,name: null == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as String,deductibility: null == deductibility ? _self.deductibility : deductibility // ignore: cast_nullable_to_non_nullable
as Deductibility,defaultDeductible: null == defaultDeductible ? _self.defaultDeductible : defaultDeductible // ignore: cast_nullable_to_non_nullable
as bool,note: freezed == note ? _self.note : note // ignore: cast_nullable_to_non_nullable
as String?,legalBasis: freezed == legalBasis ? _self.legalBasis : legalBasis // ignore: cast_nullable_to_non_nullable
as String?,defaultOperationKind: freezed == defaultOperationKind ? _self.defaultOperationKind : defaultOperationKind // ignore: cast_nullable_to_non_nullable
as OperationKind?,
  ));
}

}


/// Adds pattern-matching-related methods to [ExpenseCategory].
extension ExpenseCategoryPatterns on ExpenseCategory {
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

@optionalTypeArgs TResult maybeMap<TResult extends Object?>(TResult Function( _ExpenseCategory value)?  $default,{required TResult orElse(),}){
final _that = this;
switch (_that) {
case _ExpenseCategory() when $default != null:
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

@optionalTypeArgs TResult map<TResult extends Object?>(TResult Function( _ExpenseCategory value)  $default,){
final _that = this;
switch (_that) {
case _ExpenseCategory():
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

@optionalTypeArgs TResult? mapOrNull<TResult extends Object?>(TResult? Function( _ExpenseCategory value)?  $default,){
final _that = this;
switch (_that) {
case _ExpenseCategory() when $default != null:
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

@optionalTypeArgs TResult maybeWhen<TResult extends Object?>(TResult Function( String code,  String name,  Deductibility deductibility, @JsonKey(name: 'default_deductible')  bool defaultDeductible, @JsonKey(includeIfNull: true)  String? note, @JsonKey(includeIfNull: true, name: 'legal_basis')  String? legalBasis, @JsonKey(includeIfNull: false, name: 'default_operation_kind')  OperationKind? defaultOperationKind)?  $default,{required TResult orElse(),}) {final _that = this;
switch (_that) {
case _ExpenseCategory() when $default != null:
return $default(_that.code,_that.name,_that.deductibility,_that.defaultDeductible,_that.note,_that.legalBasis,_that.defaultOperationKind);case _:
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

@optionalTypeArgs TResult when<TResult extends Object?>(TResult Function( String code,  String name,  Deductibility deductibility, @JsonKey(name: 'default_deductible')  bool defaultDeductible, @JsonKey(includeIfNull: true)  String? note, @JsonKey(includeIfNull: true, name: 'legal_basis')  String? legalBasis, @JsonKey(includeIfNull: false, name: 'default_operation_kind')  OperationKind? defaultOperationKind)  $default,) {final _that = this;
switch (_that) {
case _ExpenseCategory():
return $default(_that.code,_that.name,_that.deductibility,_that.defaultDeductible,_that.note,_that.legalBasis,_that.defaultOperationKind);case _:
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

@optionalTypeArgs TResult? whenOrNull<TResult extends Object?>(TResult? Function( String code,  String name,  Deductibility deductibility, @JsonKey(name: 'default_deductible')  bool defaultDeductible, @JsonKey(includeIfNull: true)  String? note, @JsonKey(includeIfNull: true, name: 'legal_basis')  String? legalBasis, @JsonKey(includeIfNull: false, name: 'default_operation_kind')  OperationKind? defaultOperationKind)?  $default,) {final _that = this;
switch (_that) {
case _ExpenseCategory() when $default != null:
return $default(_that.code,_that.name,_that.deductibility,_that.defaultDeductible,_that.note,_that.legalBasis,_that.defaultOperationKind);case _:
  return null;

}
}

}

/// @nodoc
@JsonSerializable()

class _ExpenseCategory implements ExpenseCategory {
  const _ExpenseCategory({required this.code, required this.name, required this.deductibility, @JsonKey(name: 'default_deductible') required this.defaultDeductible, @JsonKey(includeIfNull: true) required this.note, @JsonKey(includeIfNull: true, name: 'legal_basis') required this.legalBasis, @JsonKey(includeIfNull: false, name: 'default_operation_kind') this.defaultOperationKind});
  factory _ExpenseCategory.fromJson(Map<String, dynamic> json) => _$ExpenseCategoryFromJson(json);

@override final  String code;
/// es-MX display name.
@override final  String name;
@override final  Deductibility deductibility;
/// The `deductible` a new expense starts with (before the CFDI and cash rules); conservative for partial cases.
@override@JsonKey(name: 'default_deductible') final  bool defaultDeductible;
/// es-MX note shown with the hint.
@override@JsonKey(includeIfNull: true) final  String? note;
/// Article-level basis (e.g. `LISR 103`).
@override@JsonKey(includeIfNull: true, name: 'legal_basis') final  String? legalBasis;
/// The `operation_kind` of this category (seed data in `category_regime_rules`, MODULE-withholdings §2.2): what a persona moral's "Tu proveedor" question E3 starts at, and what `createExpense`, `updateExpense` and `previewExpenseWithholdings` use when `operation_kind` is omitted. Clients never keep their own copy of this table. Optional in 0.10.1 (additive); sent once WH-03 seeds it and WH-07 serves it, when it **will be required**. Absent, a client leaves E3 unselected and omits `operation_kind`, so the server applies the same default.
@override@JsonKey(includeIfNull: false, name: 'default_operation_kind') final  OperationKind? defaultOperationKind;

/// Create a copy of ExpenseCategory
/// with the given fields replaced by the non-null parameter values.
@override @JsonKey(includeFromJson: false, includeToJson: false)
@pragma('vm:prefer-inline')
_$ExpenseCategoryCopyWith<_ExpenseCategory> get copyWith => __$ExpenseCategoryCopyWithImpl<_ExpenseCategory>(this, _$identity);

@override
Map<String, dynamic> toJson() {
  return _$ExpenseCategoryToJson(this, );
}

@override
bool operator ==(Object other) {
    return identical(this, other) || (other.runtimeType == runtimeType&&other is _ExpenseCategory&&(identical(other.code, code) || other.code == code)&&(identical(other.name, name) || other.name == name)&&(identical(other.deductibility, deductibility) || other.deductibility == deductibility)&&(identical(other.defaultDeductible, defaultDeductible) || other.defaultDeductible == defaultDeductible)&&(identical(other.note, note) || other.note == note)&&(identical(other.legalBasis, legalBasis) || other.legalBasis == legalBasis)&&(identical(other.defaultOperationKind, defaultOperationKind) || other.defaultOperationKind == defaultOperationKind));
}

@JsonKey(includeFromJson: false, includeToJson: false)
@override
int get hashCode {
    return Object.hash(runtimeType,code,name,deductibility,defaultDeductible,note,legalBasis,defaultOperationKind);
}

@override
String toString() {
    return 'ExpenseCategory(code: $code, name: $name, deductibility: $deductibility, defaultDeductible: $defaultDeductible, note: $note, legalBasis: $legalBasis, defaultOperationKind: $defaultOperationKind)';
}


}

/// @nodoc
abstract mixin class _$ExpenseCategoryCopyWith<$Res> implements $ExpenseCategoryCopyWith<$Res> {
  factory _$ExpenseCategoryCopyWith(_ExpenseCategory value, $Res Function(_ExpenseCategory) _then) = __$ExpenseCategoryCopyWithImpl;
@override @useResult
$Res call({
 String code, String name, Deductibility deductibility,@JsonKey(name: 'default_deductible') bool defaultDeductible,@JsonKey(includeIfNull: true) String? note,@JsonKey(includeIfNull: true, name: 'legal_basis') String? legalBasis,@JsonKey(includeIfNull: false, name: 'default_operation_kind') OperationKind? defaultOperationKind
});




}
/// @nodoc
class __$ExpenseCategoryCopyWithImpl<$Res>
    implements _$ExpenseCategoryCopyWith<$Res> {
  __$ExpenseCategoryCopyWithImpl(this._self, this._then);

  final _ExpenseCategory _self;
  final $Res Function(_ExpenseCategory) _then;

/// Create a copy of ExpenseCategory
/// with the given fields replaced by the non-null parameter values.
@override @pragma('vm:prefer-inline') $Res call({Object? code = null,Object? name = null,Object? deductibility = null,Object? defaultDeductible = null,Object? note = freezed,Object? legalBasis = freezed,Object? defaultOperationKind = freezed,}) {
  return _then(_ExpenseCategory(
code: null == code ? _self.code : code // ignore: cast_nullable_to_non_nullable
as String,name: null == name ? _self.name : name // ignore: cast_nullable_to_non_nullable
as String,deductibility: null == deductibility ? _self.deductibility : deductibility // ignore: cast_nullable_to_non_nullable
as Deductibility,defaultDeductible: null == defaultDeductible ? _self.defaultDeductible : defaultDeductible // ignore: cast_nullable_to_non_nullable
as bool,note: freezed == note ? _self.note : note // ignore: cast_nullable_to_non_nullable
as String?,legalBasis: freezed == legalBasis ? _self.legalBasis : legalBasis // ignore: cast_nullable_to_non_nullable
as String?,defaultOperationKind: freezed == defaultOperationKind ? _self.defaultOperationKind : defaultOperationKind // ignore: cast_nullable_to_non_nullable
as OperationKind?,
  ));
}


}

// dart format on
