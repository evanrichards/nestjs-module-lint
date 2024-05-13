import { Module } from "@nestjs/common";
import { SomeImport } from "./some-import";
import { OtherName as SomeOtherImport } from "src/some-other-import";
import ThirdModule from "@src/third-module";
import { ThisService, ThatService, SecretService, PrivateService } 'src/service'

@Module({
  imports: [SomeImport, SomeOtherImport],
  exports: [ThisService],
  providers: [SecretService],
})
export class TestModule {}

@Module({
  imports: [ThirdModule],
  exports: [ThatService],
  providers: [PrivateService],
})
export class OtherModule {}
