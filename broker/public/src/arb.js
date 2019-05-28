/*eslint-disable block-scoped-var, id-length, no-control-regex, no-magic-numbers, no-prototype-builtins, no-redeclare, no-shadow, no-var, sort-vars*/
"use strict";

var $protobuf = require("protobufjs/minimal");

// Common aliases
var $Reader = $protobuf.Reader, $Writer = $protobuf.Writer, $util = $protobuf.util;

// Exported root namespace
var $root = $protobuf.roots["default"] || ($protobuf.roots["default"] = {});

$root.wsapi = (function() {

    /**
     * Namespace wsapi.
     * @exports wsapi
     * @namespace
     */
    var wsapi = {};

    wsapi.ArbMarket = (function() {

        /**
         * Properties of an ArbMarket.
         * @memberof wsapi
         * @interface IArbMarket
         * @property {string|null} [hePair] ArbMarket hePair
         * @property {number|null} [spread] ArbMarket spread
         * @property {wsapi.ArbMarket.IActiveMarket|null} [low] ArbMarket low
         * @property {wsapi.ArbMarket.IActiveMarket|null} [high] ArbMarket high
         */

        /**
         * Constructs a new ArbMarket.
         * @memberof wsapi
         * @classdesc Represents an ArbMarket.
         * @implements IArbMarket
         * @constructor
         * @param {wsapi.IArbMarket=} [properties] Properties to set
         */
        function ArbMarket(properties) {
            if (properties)
                for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                    if (properties[keys[i]] != null)
                        this[keys[i]] = properties[keys[i]];
        }

        /**
         * ArbMarket hePair.
         * @member {string} hePair
         * @memberof wsapi.ArbMarket
         * @instance
         */
        ArbMarket.prototype.hePair = "";

        /**
         * ArbMarket spread.
         * @member {number} spread
         * @memberof wsapi.ArbMarket
         * @instance
         */
        ArbMarket.prototype.spread = 0;

        /**
         * ArbMarket low.
         * @member {wsapi.ArbMarket.IActiveMarket|null|undefined} low
         * @memberof wsapi.ArbMarket
         * @instance
         */
        ArbMarket.prototype.low = null;

        /**
         * ArbMarket high.
         * @member {wsapi.ArbMarket.IActiveMarket|null|undefined} high
         * @memberof wsapi.ArbMarket
         * @instance
         */
        ArbMarket.prototype.high = null;

        /**
         * Creates a new ArbMarket instance using the specified properties.
         * @function create
         * @memberof wsapi.ArbMarket
         * @static
         * @param {wsapi.IArbMarket=} [properties] Properties to set
         * @returns {wsapi.ArbMarket} ArbMarket instance
         */
        ArbMarket.create = function create(properties) {
            return new ArbMarket(properties);
        };

        /**
         * Encodes the specified ArbMarket message. Does not implicitly {@link wsapi.ArbMarket.verify|verify} messages.
         * @function encode
         * @memberof wsapi.ArbMarket
         * @static
         * @param {wsapi.IArbMarket} message ArbMarket message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ArbMarket.encode = function encode(message, writer) {
            if (!writer)
                writer = $Writer.create();
            if (message.hePair != null && message.hasOwnProperty("hePair"))
                writer.uint32(/* id 1, wireType 2 =*/10).string(message.hePair);
            if (message.spread != null && message.hasOwnProperty("spread"))
                writer.uint32(/* id 2, wireType 1 =*/17).double(message.spread);
            if (message.low != null && message.hasOwnProperty("low"))
                $root.wsapi.ArbMarket.ActiveMarket.encode(message.low, writer.uint32(/* id 3, wireType 2 =*/26).fork()).ldelim();
            if (message.high != null && message.hasOwnProperty("high"))
                $root.wsapi.ArbMarket.ActiveMarket.encode(message.high, writer.uint32(/* id 4, wireType 2 =*/34).fork()).ldelim();
            return writer;
        };

        /**
         * Encodes the specified ArbMarket message, length delimited. Does not implicitly {@link wsapi.ArbMarket.verify|verify} messages.
         * @function encodeDelimited
         * @memberof wsapi.ArbMarket
         * @static
         * @param {wsapi.IArbMarket} message ArbMarket message or plain object to encode
         * @param {$protobuf.Writer} [writer] Writer to encode to
         * @returns {$protobuf.Writer} Writer
         */
        ArbMarket.encodeDelimited = function encodeDelimited(message, writer) {
            return this.encode(message, writer).ldelim();
        };

        /**
         * Decodes an ArbMarket message from the specified reader or buffer.
         * @function decode
         * @memberof wsapi.ArbMarket
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @param {number} [length] Message length if known beforehand
         * @returns {wsapi.ArbMarket} ArbMarket
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ArbMarket.decode = function decode(reader, length) {
            if (!(reader instanceof $Reader))
                reader = $Reader.create(reader);
            var end = length === undefined ? reader.len : reader.pos + length, message = new $root.wsapi.ArbMarket();
            while (reader.pos < end) {
                var tag = reader.uint32();
                switch (tag >>> 3) {
                case 1:
                    message.hePair = reader.string();
                    break;
                case 2:
                    message.spread = reader.double();
                    break;
                case 3:
                    message.low = $root.wsapi.ArbMarket.ActiveMarket.decode(reader, reader.uint32());
                    break;
                case 4:
                    message.high = $root.wsapi.ArbMarket.ActiveMarket.decode(reader, reader.uint32());
                    break;
                default:
                    reader.skipType(tag & 7);
                    break;
                }
            }
            return message;
        };

        /**
         * Decodes an ArbMarket message from the specified reader or buffer, length delimited.
         * @function decodeDelimited
         * @memberof wsapi.ArbMarket
         * @static
         * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
         * @returns {wsapi.ArbMarket} ArbMarket
         * @throws {Error} If the payload is not a reader or valid buffer
         * @throws {$protobuf.util.ProtocolError} If required fields are missing
         */
        ArbMarket.decodeDelimited = function decodeDelimited(reader) {
            if (!(reader instanceof $Reader))
                reader = new $Reader(reader);
            return this.decode(reader, reader.uint32());
        };

        /**
         * Verifies an ArbMarket message.
         * @function verify
         * @memberof wsapi.ArbMarket
         * @static
         * @param {Object.<string,*>} message Plain object to verify
         * @returns {string|null} `null` if valid, otherwise the reason why it is not
         */
        ArbMarket.verify = function verify(message) {
            if (typeof message !== "object" || message === null)
                return "object expected";
            if (message.hePair != null && message.hasOwnProperty("hePair"))
                if (!$util.isString(message.hePair))
                    return "hePair: string expected";
            if (message.spread != null && message.hasOwnProperty("spread"))
                if (typeof message.spread !== "number")
                    return "spread: number expected";
            if (message.low != null && message.hasOwnProperty("low")) {
                var error = $root.wsapi.ArbMarket.ActiveMarket.verify(message.low);
                if (error)
                    return "low." + error;
            }
            if (message.high != null && message.hasOwnProperty("high")) {
                var error = $root.wsapi.ArbMarket.ActiveMarket.verify(message.high);
                if (error)
                    return "high." + error;
            }
            return null;
        };

        /**
         * Creates an ArbMarket message from a plain object. Also converts values to their respective internal types.
         * @function fromObject
         * @memberof wsapi.ArbMarket
         * @static
         * @param {Object.<string,*>} object Plain object
         * @returns {wsapi.ArbMarket} ArbMarket
         */
        ArbMarket.fromObject = function fromObject(object) {
            if (object instanceof $root.wsapi.ArbMarket)
                return object;
            var message = new $root.wsapi.ArbMarket();
            if (object.hePair != null)
                message.hePair = String(object.hePair);
            if (object.spread != null)
                message.spread = Number(object.spread);
            if (object.low != null) {
                if (typeof object.low !== "object")
                    throw TypeError(".wsapi.ArbMarket.low: object expected");
                message.low = $root.wsapi.ArbMarket.ActiveMarket.fromObject(object.low);
            }
            if (object.high != null) {
                if (typeof object.high !== "object")
                    throw TypeError(".wsapi.ArbMarket.high: object expected");
                message.high = $root.wsapi.ArbMarket.ActiveMarket.fromObject(object.high);
            }
            return message;
        };

        /**
         * Creates a plain object from an ArbMarket message. Also converts values to other types if specified.
         * @function toObject
         * @memberof wsapi.ArbMarket
         * @static
         * @param {wsapi.ArbMarket} message ArbMarket
         * @param {$protobuf.IConversionOptions} [options] Conversion options
         * @returns {Object.<string,*>} Plain object
         */
        ArbMarket.toObject = function toObject(message, options) {
            if (!options)
                options = {};
            var object = {};
            if (options.defaults) {
                object.hePair = "";
                object.spread = 0;
                object.low = null;
                object.high = null;
            }
            if (message.hePair != null && message.hasOwnProperty("hePair"))
                object.hePair = message.hePair;
            if (message.spread != null && message.hasOwnProperty("spread"))
                object.spread = options.json && !isFinite(message.spread) ? String(message.spread) : message.spread;
            if (message.low != null && message.hasOwnProperty("low"))
                object.low = $root.wsapi.ArbMarket.ActiveMarket.toObject(message.low, options);
            if (message.high != null && message.hasOwnProperty("high"))
                object.high = $root.wsapi.ArbMarket.ActiveMarket.toObject(message.high, options);
            return object;
        };

        /**
         * Converts this ArbMarket to JSON.
         * @function toJSON
         * @memberof wsapi.ArbMarket
         * @instance
         * @returns {Object.<string,*>} JSON object
         */
        ArbMarket.prototype.toJSON = function toJSON() {
            return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
        };

        ArbMarket.ActiveMarket = (function() {

            /**
             * Properties of an ActiveMarket.
             * @memberof wsapi.ArbMarket
             * @interface IActiveMarket
             * @property {string|null} [exchange] ActiveMarket exchange
             * @property {string|null} [hePair] ActiveMarket hePair
             * @property {string|null} [exPair] ActiveMarket exPair
             * @property {string|null} [price] ActiveMarket price
             */

            /**
             * Constructs a new ActiveMarket.
             * @memberof wsapi.ArbMarket
             * @classdesc Represents an ActiveMarket.
             * @implements IActiveMarket
             * @constructor
             * @param {wsapi.ArbMarket.IActiveMarket=} [properties] Properties to set
             */
            function ActiveMarket(properties) {
                if (properties)
                    for (var keys = Object.keys(properties), i = 0; i < keys.length; ++i)
                        if (properties[keys[i]] != null)
                            this[keys[i]] = properties[keys[i]];
            }

            /**
             * ActiveMarket exchange.
             * @member {string} exchange
             * @memberof wsapi.ArbMarket.ActiveMarket
             * @instance
             */
            ActiveMarket.prototype.exchange = "";

            /**
             * ActiveMarket hePair.
             * @member {string} hePair
             * @memberof wsapi.ArbMarket.ActiveMarket
             * @instance
             */
            ActiveMarket.prototype.hePair = "";

            /**
             * ActiveMarket exPair.
             * @member {string} exPair
             * @memberof wsapi.ArbMarket.ActiveMarket
             * @instance
             */
            ActiveMarket.prototype.exPair = "";

            /**
             * ActiveMarket price.
             * @member {string} price
             * @memberof wsapi.ArbMarket.ActiveMarket
             * @instance
             */
            ActiveMarket.prototype.price = "";

            /**
             * Creates a new ActiveMarket instance using the specified properties.
             * @function create
             * @memberof wsapi.ArbMarket.ActiveMarket
             * @static
             * @param {wsapi.ArbMarket.IActiveMarket=} [properties] Properties to set
             * @returns {wsapi.ArbMarket.ActiveMarket} ActiveMarket instance
             */
            ActiveMarket.create = function create(properties) {
                return new ActiveMarket(properties);
            };

            /**
             * Encodes the specified ActiveMarket message. Does not implicitly {@link wsapi.ArbMarket.ActiveMarket.verify|verify} messages.
             * @function encode
             * @memberof wsapi.ArbMarket.ActiveMarket
             * @static
             * @param {wsapi.ArbMarket.IActiveMarket} message ActiveMarket message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            ActiveMarket.encode = function encode(message, writer) {
                if (!writer)
                    writer = $Writer.create();
                if (message.exchange != null && message.hasOwnProperty("exchange"))
                    writer.uint32(/* id 1, wireType 2 =*/10).string(message.exchange);
                if (message.hePair != null && message.hasOwnProperty("hePair"))
                    writer.uint32(/* id 2, wireType 2 =*/18).string(message.hePair);
                if (message.exPair != null && message.hasOwnProperty("exPair"))
                    writer.uint32(/* id 3, wireType 2 =*/26).string(message.exPair);
                if (message.price != null && message.hasOwnProperty("price"))
                    writer.uint32(/* id 4, wireType 2 =*/34).string(message.price);
                return writer;
            };

            /**
             * Encodes the specified ActiveMarket message, length delimited. Does not implicitly {@link wsapi.ArbMarket.ActiveMarket.verify|verify} messages.
             * @function encodeDelimited
             * @memberof wsapi.ArbMarket.ActiveMarket
             * @static
             * @param {wsapi.ArbMarket.IActiveMarket} message ActiveMarket message or plain object to encode
             * @param {$protobuf.Writer} [writer] Writer to encode to
             * @returns {$protobuf.Writer} Writer
             */
            ActiveMarket.encodeDelimited = function encodeDelimited(message, writer) {
                return this.encode(message, writer).ldelim();
            };

            /**
             * Decodes an ActiveMarket message from the specified reader or buffer.
             * @function decode
             * @memberof wsapi.ArbMarket.ActiveMarket
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @param {number} [length] Message length if known beforehand
             * @returns {wsapi.ArbMarket.ActiveMarket} ActiveMarket
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            ActiveMarket.decode = function decode(reader, length) {
                if (!(reader instanceof $Reader))
                    reader = $Reader.create(reader);
                var end = length === undefined ? reader.len : reader.pos + length, message = new $root.wsapi.ArbMarket.ActiveMarket();
                while (reader.pos < end) {
                    var tag = reader.uint32();
                    switch (tag >>> 3) {
                    case 1:
                        message.exchange = reader.string();
                        break;
                    case 2:
                        message.hePair = reader.string();
                        break;
                    case 3:
                        message.exPair = reader.string();
                        break;
                    case 4:
                        message.price = reader.string();
                        break;
                    default:
                        reader.skipType(tag & 7);
                        break;
                    }
                }
                return message;
            };

            /**
             * Decodes an ActiveMarket message from the specified reader or buffer, length delimited.
             * @function decodeDelimited
             * @memberof wsapi.ArbMarket.ActiveMarket
             * @static
             * @param {$protobuf.Reader|Uint8Array} reader Reader or buffer to decode from
             * @returns {wsapi.ArbMarket.ActiveMarket} ActiveMarket
             * @throws {Error} If the payload is not a reader or valid buffer
             * @throws {$protobuf.util.ProtocolError} If required fields are missing
             */
            ActiveMarket.decodeDelimited = function decodeDelimited(reader) {
                if (!(reader instanceof $Reader))
                    reader = new $Reader(reader);
                return this.decode(reader, reader.uint32());
            };

            /**
             * Verifies an ActiveMarket message.
             * @function verify
             * @memberof wsapi.ArbMarket.ActiveMarket
             * @static
             * @param {Object.<string,*>} message Plain object to verify
             * @returns {string|null} `null` if valid, otherwise the reason why it is not
             */
            ActiveMarket.verify = function verify(message) {
                if (typeof message !== "object" || message === null)
                    return "object expected";
                if (message.exchange != null && message.hasOwnProperty("exchange"))
                    if (!$util.isString(message.exchange))
                        return "exchange: string expected";
                if (message.hePair != null && message.hasOwnProperty("hePair"))
                    if (!$util.isString(message.hePair))
                        return "hePair: string expected";
                if (message.exPair != null && message.hasOwnProperty("exPair"))
                    if (!$util.isString(message.exPair))
                        return "exPair: string expected";
                if (message.price != null && message.hasOwnProperty("price"))
                    if (!$util.isString(message.price))
                        return "price: string expected";
                return null;
            };

            /**
             * Creates an ActiveMarket message from a plain object. Also converts values to their respective internal types.
             * @function fromObject
             * @memberof wsapi.ArbMarket.ActiveMarket
             * @static
             * @param {Object.<string,*>} object Plain object
             * @returns {wsapi.ArbMarket.ActiveMarket} ActiveMarket
             */
            ActiveMarket.fromObject = function fromObject(object) {
                if (object instanceof $root.wsapi.ArbMarket.ActiveMarket)
                    return object;
                var message = new $root.wsapi.ArbMarket.ActiveMarket();
                if (object.exchange != null)
                    message.exchange = String(object.exchange);
                if (object.hePair != null)
                    message.hePair = String(object.hePair);
                if (object.exPair != null)
                    message.exPair = String(object.exPair);
                if (object.price != null)
                    message.price = String(object.price);
                return message;
            };

            /**
             * Creates a plain object from an ActiveMarket message. Also converts values to other types if specified.
             * @function toObject
             * @memberof wsapi.ArbMarket.ActiveMarket
             * @static
             * @param {wsapi.ArbMarket.ActiveMarket} message ActiveMarket
             * @param {$protobuf.IConversionOptions} [options] Conversion options
             * @returns {Object.<string,*>} Plain object
             */
            ActiveMarket.toObject = function toObject(message, options) {
                if (!options)
                    options = {};
                var object = {};
                if (options.defaults) {
                    object.exchange = "";
                    object.hePair = "";
                    object.exPair = "";
                    object.price = "";
                }
                if (message.exchange != null && message.hasOwnProperty("exchange"))
                    object.exchange = message.exchange;
                if (message.hePair != null && message.hasOwnProperty("hePair"))
                    object.hePair = message.hePair;
                if (message.exPair != null && message.hasOwnProperty("exPair"))
                    object.exPair = message.exPair;
                if (message.price != null && message.hasOwnProperty("price"))
                    object.price = message.price;
                return object;
            };

            /**
             * Converts this ActiveMarket to JSON.
             * @function toJSON
             * @memberof wsapi.ArbMarket.ActiveMarket
             * @instance
             * @returns {Object.<string,*>} JSON object
             */
            ActiveMarket.prototype.toJSON = function toJSON() {
                return this.constructor.toObject(this, $protobuf.util.toJSONOptions);
            };

            return ActiveMarket;
        })();

        return ArbMarket;
    })();

    return wsapi;
})();

module.exports = $root;
