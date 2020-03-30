module Proof = {
  type t = {
    jsonProof: Js.Json.t,
    evmProofBytes: JsBuffer.t,
  };

  let decodeProof = json =>
    JsonUtils.Decode.{
      jsonProof: json |> at(["result", "jsonProof"], json => json),
      evmProofBytes: json |> at(["result", "evmProofBytes"], string) |> JsBuffer.fromHex,
    };
};

let get = (requestId: int) => {
  let json = AxiosHooks.use("bandchain/proof/" ++ (requestId |> string_of_int));
  json |> Belt.Option.map(_, Proof.decodeProof);
};
