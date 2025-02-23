target triple = "x86_64-pc-windows-msvc19.43.34808"

@str.0 = global [3 x i8] c"%s\00"
@str.1 = global [5 x i8] c"%lld\00"
@str.2 = global [9 x i8] c"%255[^\0A]\00"
@str.3 = global [4 x i8] c"%*c\00"
@str.4 = global [4 x i8] c"a: \00"
@str.5 = global [2 x i8] c"\0A\00"
@str.6 = global [4 x i8] c"b: \00"
@str.7 = global [9 x i8] c"a / b = \00"
@str.8 = global [21 x i8] c" (integer division)\0A\00"
@str.9 = global [9 x i8] c"a + b = \00"
@str.10 = global [9 x i8] c"a * b = \00"
@str.11 = global [9 x i8] c"b - a = \00"
@str.12 = global [27 x i8] c"testing pratt parsing for\0A\00"
@str.13 = global [35 x i8] c"1 - 2 + 3 * 4 + a * b - (2 + 3) = \00"

declare i32 @printf(i8* %format, ...)

declare i32 @scanf(i8* %format, ...)

declare i8* @memset(i8* %str, i32 %c, i64 %n)

define i8* @IO_out_string(i8* %self, i8* %x) {
0:
	%1 = call i32 (i8*, ...) @printf(i8* getelementptr ([3 x i8], [3 x i8]* @str.0, i32 0, i32 0), i8* %x)
	ret i8* %self
}

define i8* @IO_out_int(i8* %self, i64 %x) {
0:
	%1 = call i32 (i8*, ...) @printf(i8* getelementptr ([5 x i8], [5 x i8]* @str.1, i32 0, i32 0), i64 %x)
	ret i8* %self
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
	%1 = alloca i64
	store i64 15, i64* %1
	%2 = alloca i64
	store i64 4, i64* %2
	%3 = call i8* @IO_out_string(i8* null, i8* getelementptr ([4 x i8], [4 x i8]* @str.4, i32 0, i32 0))
	%4 = load i64, i64* %1
	%5 = call i8* @IO_out_int(i8* null, i64 %4)
	%6 = call i8* @IO_out_string(i8* null, i8* getelementptr ([2 x i8], [2 x i8]* @str.5, i32 0, i32 0))
	%7 = call i8* @IO_out_string(i8* null, i8* getelementptr ([4 x i8], [4 x i8]* @str.6, i32 0, i32 0))
	%8 = load i64, i64* %2
	%9 = call i8* @IO_out_int(i8* null, i64 %8)
	%10 = call i8* @IO_out_string(i8* null, [2 x i8]* @str.5)
	%11 = call i8* @IO_out_string(i8* null, i8* getelementptr ([9 x i8], [9 x i8]* @str.7, i32 0, i32 0))
	%12 = load i64, i64* %1
	%13 = load i64, i64* %2
	%14 = sdiv i64 %12, %13
	%15 = call i8* @IO_out_int(i8* null, i64 %14)
	%16 = call i8* @IO_out_string(i8* null, i8* getelementptr ([21 x i8], [21 x i8]* @str.8, i32 0, i32 0))
	%17 = call i8* @IO_out_string(i8* null, i8* getelementptr ([9 x i8], [9 x i8]* @str.9, i32 0, i32 0))
	%18 = load i64, i64* %1
	%19 = load i64, i64* %2
	%20 = add i64 %18, %19
	%21 = call i8* @IO_out_int(i8* null, i64 %20)
	%22 = call i8* @IO_out_string(i8* null, [2 x i8]* @str.5)
	%23 = call i8* @IO_out_string(i8* null, i8* getelementptr ([9 x i8], [9 x i8]* @str.10, i32 0, i32 0))
	%24 = load i64, i64* %1
	%25 = load i64, i64* %2
	%26 = mul i64 %24, %25
	%27 = call i8* @IO_out_int(i8* null, i64 %26)
	%28 = call i8* @IO_out_string(i8* null, [2 x i8]* @str.5)
	%29 = call i8* @IO_out_string(i8* null, i8* getelementptr ([9 x i8], [9 x i8]* @str.11, i32 0, i32 0))
	%30 = load i64, i64* %2
	%31 = load i64, i64* %1
	%32 = sub i64 %30, %31
	%33 = call i8* @IO_out_int(i8* null, i64 %32)
	%34 = call i8* @IO_out_string(i8* null, [2 x i8]* @str.5)
	%35 = call i8* @IO_out_string(i8* null, i8* getelementptr ([27 x i8], [27 x i8]* @str.12, i32 0, i32 0))
	%36 = call i8* @IO_out_string(i8* null, i8* getelementptr ([35 x i8], [35 x i8]* @str.13, i32 0, i32 0))
	%37 = sub i64 1, 2
	%38 = mul i64 3, 4
	%39 = add i64 %37, %38
	%40 = load i64, i64* %1
	%41 = load i64, i64* %2
	%42 = mul i64 %40, %41
	%43 = add i64 %39, %42
	%44 = add i64 2, 3
	%45 = sub i64 %43, %44
	%46 = call i8* @IO_out_int(i8* null, i64 %45)
	ret i8* %46
}

define i32 @main() {
0:
	%1 = call i8* @Main_main(i8* null)
	ret i32 0
}
