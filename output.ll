target triple = "x86_64-pc-windows-msvc"

@str.0 = global [3 x i8] c"%s\00"
@str.1 = global [5 x i8] c"%lld\00"
@str.2 = global [9 x i8] c"%255[^\0A]\00"
@str.3 = global [4 x i8] c"%*c\00"
@str.4 = global [6 x i8] c"true\0A\00"
@str.5 = global [7 x i8] c"false\0A\00"

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
	%1 = icmp slt i64 2, 1
	%2 = icmp slt i64 0, 1
	br i1 %2, label %7, label %8

3:
	call void @IO_out_string(i8* null, i8* getelementptr ([6 x i8], [6 x i8]* @str.4, i32 0, i32 0))
	br label %5

4:
	call void @IO_out_string(i8* null, i8* getelementptr ([7 x i8], [7 x i8]* @str.5, i32 0, i32 0))
	br label %5

5:
	%6 = phi i8* [ null, %3 ], [ null, %4 ]
	br label %5

7:
	call void @IO_out_string(i8* null, [6 x i8]* @str.4)
	br label %9

8:
	call void @IO_out_string(i8* null, [7 x i8]* @str.5)
	br label %9

9:
	%10 = phi i8* [ null, %7 ], [ null, %8 ]
	br label %9
}

define i32 @main() {
0:
	%1 = call i8* @Main_main(i8* null)
	ret i32 0
}
