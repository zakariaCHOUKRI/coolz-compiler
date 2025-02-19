target triple = "x86_64-pc-windows-msvc"

@str.0 = global [3 x i8] c"%s\00"
@str.1 = global [5 x i8] c"%lld\00"
@str.2 = global [13 x i8] c"Hello World\0A\00"
@str.3 = global [20 x i8] c"My name is Zakaria\0A\00"
@str.4 = global [18 x i8] c"Nice to meet you\0A\00"
@str.5 = global [16 x i8] c"test test test\0A\00"
@str.6 = global [15 x i8] c"amine idrissi\0A\00"

declare i32 @printf(i8* %format, ...)

define void @IO_out_string(i8* %self, i8* %x) {
0:
	%1 = call i32 (i8*, ...) @printf(i8* getelementptr ([3 x i8], [3 x i8]* @str.0, i32 0, i32 0), i8* %x)
	ret void
}

define void @IO_out_int(i8* %self, i64 %x) {
0:
	%1 = call i32 (i8*, ...) @printf(i8* getelementptr ([5 x i8], [5 x i8]* @str.1, i32 0, i32 0), i64 %x)
	ret void
}

define i8* @Main_main(i8* %self) {
0:
	call void @IO_out_string(i8* null, i8* getelementptr ([13 x i8], [13 x i8]* @str.2, i32 0, i32 0))
	call void @IO_out_string(i8* null, i8* getelementptr ([20 x i8], [20 x i8]* @str.3, i32 0, i32 0))
	call void @IO_out_string(i8* null, i8* getelementptr ([18 x i8], [18 x i8]* @str.4, i32 0, i32 0))
	call void @IO_out_string(i8* null, i8* getelementptr ([16 x i8], [16 x i8]* @str.5, i32 0, i32 0))
	call void @IO_out_string(i8* null, i8* getelementptr ([15 x i8], [15 x i8]* @str.6, i32 0, i32 0))
	call void @IO_out_int(i8* null, i64 65465465454)
	ret i8* null
}

define i32 @main() {
0:
	%1 = call i8* @Main_main(i8* null)
	ret i32 0
}
