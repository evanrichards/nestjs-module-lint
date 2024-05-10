import { Module } from "@nestjs/common";

@Module({
  imports: [SomeImport, SomeOtherImport],
  exports: [],
  providers: [],
})
export class TestModule {}

@Module({
  imports: [ThirdModule],
  exports: [],
  providers: [],
})
export class OtherModule {}
