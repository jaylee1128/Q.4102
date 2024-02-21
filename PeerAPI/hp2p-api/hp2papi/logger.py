from datetime import datetime
import logging

logger = logging.getLogger("hp2p.api")
logger.setLevel(logging.DEBUG)

peerlogger = logging.getLogger("hp2p.peer")
peerlogger.setLevel(logging.DEBUG)

ch = logging.StreamHandler()

formatter = logging.Formatter(
    "[%(levelname)s][%(name)s] %(asctime)s %(message)s"
)
ch.setFormatter(formatter)

# add ch to logger
logger.addHandler(ch)
peerlogger.addHandler(ch)

def SetLogLevel(level: str):
    """
    로그 레벨 설정
        level: "DEBUG", "INFO", "WARNING", "ERROR", "CRITICAL"
    """
    global logger, peerlogger

    if level == "DEBUG":
        logger.setLevel(logging.DEBUG)
        peerlogger.setLevel(logging.DEBUG)
    elif level == "INFO":
        logger.setLevel(logging.INFO)
        peerlogger.setLevel(logging.INFO)
    elif level == "WARNING":
        logger.setLevel(logging.WARNING)
        peerlogger.setLevel(logging.WARNING)
    elif level == "ERROR":
        logger.setLevel(logging.ERROR)
        peerlogger.setLevel(logging.ERROR)
    elif level == "CRITICAL":
        logger.setLevel(logging.CRITICAL)
        peerlogger.setLevel(logging.CRITICAL)
    else:
        logger.setLevel(logging.INFO)
        peerlogger.setLevel(logging.INFO)

def SetLogToFile():
    """
    로그 파일로 출력
    """
    global logger, formatter

    current = datetime.now()
    filename = "hp2papi_" + current.strftime("%Y-%m-%d_%H_%M_%S") + ".log"
    fh = logging.FileHandler(filename=filename, encoding="utf-8")
    fh.setFormatter(formatter)

    logger.handlers.clear()
    logger.addHandler(fh)
    peerlogger.handlers.clear()
    peerlogger.addHandler(fh)

def print(*values: object) -> None:
    printInfo(*values)

def printDebug(*values: object) -> None:
    global logger
    logger.debug(*values)

def printInfo(*values: object) -> None:
    global logger
    logger.info(*values)

def printWarning(*values: object) -> None:
    global logger
    logger.warning(*values)

def printError(*values: object) -> None:
    global logger
    logger.error(*values)

def printCritical(*values: object) -> None:
    global logger
    logger.critical(*values)

def peerprint(*values: object) -> None:
    global peerlogger
    peerlogger.debug(*values)
