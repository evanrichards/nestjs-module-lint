import { Module } from "@nestjs/common";
import { SomeImport } from "./some-import";
import { OtherName as SomeOtherImport } from "src/some-other-import";
import ThirdModule from "@src/third-module";

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
