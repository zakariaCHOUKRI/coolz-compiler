target triple = "x86_64-pc-windows-msvc"

@str.0 = global [3 x i8] c"%s\00"
@str.1 = global [5 x i8] c"%lld\00"
@str.2 = global [9 x i8] c"%255[^\0A]\00"
@str.3 = global [4 x i8] c"%*c\00"
@str.4 = global [25 x i8] c"Please enter your name: \00"
@str.5 = global [15 x i8] c"Your name is: \00"
@str.6 = global [2 x i8] c"\0A\00"
@str.7 = global [24 x i8] c"Please enter a number: \00"
@str.8 = global [17 x i8] c"Your number is: \00"

declare i32 @printf(i8* %format, ...)

declare i32 @scanf(i8* %format, ...)

declare i8* @memset(i8* %str, i32 %c, i64 %n)

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

define i8* @IO_in_string(i8* %self) {
0:
	%1 = alloca [256 x i8]
	%2 = getelementptr [256 x i8], [256 x i8]* %1, i32 0, i32 0
	%3 = call i8* @memset(i8* %2, i32 0, i64 256)
	%4 = call i32 (i8*, ...) @scanf(i8* getelementptr ([9 x i8], [9 x i8]* @str.2, i32 0, i32 0), [256 x i8]* %1)
	%5 = getelementptr [256 x i8], [256 x i8]* %1, i32 0, i32 0
	%6 = call i64 @strlen(i8* %5)
	%7 = add i64 %6, 1
	%8 = call i8* @malloc(i64 %7)
	%9 = call i8* @memcpy(i8* %8, i8* %5, i64 %7)
	ret i8* %8
}

declare i64 @strlen(i8* %str)

declare i8* @malloc(i64 %size)

declare i8* @memcpy(i8* %dest, i8* %src, i64 %size)

define i64 @IO_in_int(i8* %self) {
0:
	%1 = alloca i64
	%2 = call i32 (i8*, ...) @scanf([5 x i8]* @str.1, i64* %1)
	%3 = call i32 (i8*, ...) @scanf(i8* getelementptr ([4 x i8], [4 x i8]* @str.3, i32 0, i32 0))
	%4 = load i64, i64* %1
	ret i64 %4
}

define i8* @Main_main(i8* %self) {
0:
	call void @IO_out_string(i8* null, i8* getelementptr ([25 x i8], [25 x i8]* @str.4, i32 0, i32 0))
	%1 = alloca i8*
	%2 = call i8* @IO_in_string(i8* null)
	store i8* %2, i8** %1
	call void @IO_out_string(i8* null, i8* getelementptr ([15 x i8], [15 x i8]* @str.5, i32 0, i32 0))
	%3 = load i8*, i8** %1
	call void @IO_out_string(i8* null, i8* %3)
	call void @IO_out_string(i8* null, i8* getelementptr ([2 x i8], [2 x i8]* @str.6, i32 0, i32 0))
	call void @IO_out_string(i8* null, i8* getelementptr ([24 x i8], [24 x i8]* @str.7, i32 0, i32 0))
	%4 = alloca i64
	%5 = call i64 @IO_in_int(i8* null)
	store i64 %5, i64* %4
	call void @IO_out_string(i8* null, i8* getelementptr ([17 x i8], [17 x i8]* @str.8, i32 0, i32 0))
	%6 = load i64, i64* %4
	call void @IO_out_int(i8* null, i64 %6)
	call void @IO_out_string(i8* null, [2 x i8]* @str.6)
	call void @IO_out_string(i8* null, [24 x i8]* @str.7)
	%7 = alloca i64
	%8 = call i64 @IO_in_int(i8* null)
	store i64 %8, i64* %7
	call void @IO_out_string(i8* null, [17 x i8]* @str.8)
	%9 = load i64, i64* %7
	call void @IO_out_int(i8* null, i64 %9)
	call void @IO_out_string(i8* null, [2 x i8]* @str.6)
	call void @IO_out_string(i8* null, [25 x i8]* @str.4)
	%10 = alloca i8*
	%11 = call i8* @IO_in_string(i8* null)
	store i8* %11, i8** %10
	call void @IO_out_string(i8* null, [15 x i8]* @str.5)
	%12 = load i8*, i8** %10
	call void @IO_out_string(i8* null, i8* %12)
	call void @IO_out_string(i8* null, [2 x i8]* @str.6)
	ret i8* null
}

define i32 @main() {
0:
	%1 = call i8* @Main_main(i8* null)
	ret i32 0
}
